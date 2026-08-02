#!/usr/bin/env python3
"""Validate all SQL migrations against a temporary PostgreSQL container."""

from __future__ import annotations

from pathlib import Path
import os
import subprocess
import sys
import time
from typing import Final
from uuid import uuid4

POSTGRES_IMAGE: Final = "postgres:18.4-bookworm"
DATABASE_NAME: Final = "groundedsearch"
DATABASE_USER: Final = "groundedsearch"
DATABASE_PASSWORD: Final = "groundedsearch_dev"
STARTUP_TIMEOUT_SECONDS: Final = 60.0

EXPECTED_TABLES: Final = {
    "document_acl_groups",
    "document_acl_users",
    "document_chunks",
    "documents",
    "group_memberships",
    "groups",
    "users",
}


class MigrationValidationError(RuntimeError):
    """Raised when migration validation fails."""


def run(
    command: list[str],
    *,
    cwd: Path,
    input_text: str | None = None,
    capture: bool = True,
) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(
        command,
        cwd=cwd,
        input=input_text,
        text=True,
        capture_output=capture,
        check=False,
    )
    if result.returncode != 0:
        raise MigrationValidationError(
            f"command failed ({result.returncode}): {' '.join(command)}\n"
            f"{result.stdout or ''}{result.stderr or ''}"
        )
    return result


def repository_root() -> Path:
    result = subprocess.run(
        ["git", "rev-parse", "--show-toplevel"],
        text=True,
        capture_output=True,
        check=False,
    )
    if result.returncode != 0:
        raise MigrationValidationError(
            "run this script from inside the GroundedSearch repository"
        )
    return Path(result.stdout.strip()).resolve()


def migration_files(root: Path, direction: str) -> list[Path]:
    files = sorted((root / "db" / "migrations").glob(f"*.{direction}.sql"))
    if not files:
        raise MigrationValidationError(
            f"no {direction!r} migration files were found"
        )
    if direction == "down":
        files.reverse()
    return files


def docker_exec_psql(
    root: Path,
    container_name: str,
    *,
    sql: str | None = None,
    query: str | None = None,
) -> str:
    command = [
        "docker",
        "exec",
        "-i",
        container_name,
        "psql",
        "--set",
        "ON_ERROR_STOP=1",
        "--username",
        DATABASE_USER,
        "--dbname",
        DATABASE_NAME,
        "--no-psqlrc",
    ]

    if query is not None:
        command.extend(["--tuples-only", "--no-align", "--command", query])

    result = run(
        command,
        cwd=root,
        input_text=sql,
    )
    return result.stdout.strip()


def wait_for_postgres(root: Path, container_name: str) -> None:
    deadline = time.monotonic() + STARTUP_TIMEOUT_SECONDS
    last_output = ""

    while time.monotonic() < deadline:
        result = subprocess.run(
            [
                "docker",
                "exec",
                container_name,
                "pg_isready",
                "--username",
                DATABASE_USER,
                "--dbname",
                DATABASE_NAME,
            ],
            cwd=root,
            text=True,
            capture_output=True,
            check=False,
        )
        last_output = f"{result.stdout}{result.stderr}".strip()
        if result.returncode == 0:
            return
        time.sleep(1.0)

    raise MigrationValidationError(
        "PostgreSQL did not become ready before the timeout: "
        f"{last_output}"
    )


def current_tables(root: Path, container_name: str) -> set[str]:
    output = docker_exec_psql(
        root,
        container_name,
        query=(
            "SELECT tablename "
            "FROM pg_catalog.pg_tables "
            "WHERE schemaname = 'public' "
            "ORDER BY tablename;"
        ),
    )
    return {line for line in output.splitlines() if line}


def apply_migrations(
    root: Path,
    container_name: str,
    files: list[Path],
) -> None:
    for path in files:
        print(f"Applying {path.name}")
        docker_exec_psql(
            root,
            container_name,
            sql=path.read_text(encoding="utf-8"),
        )


def validate_constraints(root: Path, container_name: str) -> None:
    constraint_count = docker_exec_psql(
        root,
        container_name,
        query=(
            "SELECT count(*) "
            "FROM pg_catalog.pg_constraint "
            "WHERE conname IN ("
            "'documents_status_valid',"
            "'document_chunks_offsets_valid',"
            "'document_acl_users_decision_valid',"
            "'document_acl_groups_decision_valid'"
            ");"
        ),
    )
    if constraint_count != "4":
        raise MigrationValidationError(
            "expected four critical constraints, "
            f"found {constraint_count!r}"
        )

    docker_exec_psql(
        root,
        container_name,
        sql="""
INSERT INTO users (
    id,
    external_id,
    display_name,
    email
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'alice',
    'Alice Example',
    'alice@example.test'
);

INSERT INTO groups (
    id,
    external_id,
    display_name
) VALUES (
    '00000000-0000-0000-0000-000000000101',
    'engineering',
    'Engineering'
);

INSERT INTO group_memberships (
    user_id,
    group_id
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000101'
);

INSERT INTO documents (
    id,
    source_type,
    source_id,
    title,
    content_hash,
    is_public
) VALUES (
    '00000000-0000-0000-0000-000000001001',
    'fixture',
    'public-release-notes',
    'Public release notes',
    'sha256:fixture',
    true
);

INSERT INTO document_chunks (
    id,
    document_id,
    ordinal,
    content,
    content_hash,
    start_offset,
    end_offset
) VALUES (
    '00000000-0000-0000-0000-000000002001',
    '00000000-0000-0000-0000-000000001001',
    0,
    'GroundedSearch fixture content.',
    'sha256:chunk-fixture',
    0,
    31
);

INSERT INTO document_acl_users (
    document_id,
    user_id,
    decision
) VALUES (
    '00000000-0000-0000-0000-000000001001',
    '00000000-0000-0000-0000-000000000001',
    'deny'
);
""",
    )


def main() -> int:
    root = repository_root()
    container_name = f"grounded-search-migrations-{os.getpid()}-{uuid4().hex[:8]}"

    run(["docker", "version"], cwd=root)

    try:
        print(f"Starting PostgreSQL container: {container_name}")
        run(
            [
                "docker",
                "run",
                "--detach",
                "--name",
                container_name,
                "--env",
                f"POSTGRES_USER={DATABASE_USER}",
                "--env",
                f"POSTGRES_PASSWORD={DATABASE_PASSWORD}",
                "--env",
                f"POSTGRES_DB={DATABASE_NAME}",
                POSTGRES_IMAGE,
            ],
            cwd=root,
            capture=False,
        )

        wait_for_postgres(root, container_name)

        up_files = migration_files(root, "up")
        down_files = migration_files(root, "down")

        if len(up_files) != len(down_files):
            raise MigrationValidationError(
                "up/down migration counts do not match"
            )

        apply_migrations(root, container_name, up_files)

        tables = current_tables(root, container_name)
        if tables != EXPECTED_TABLES:
            raise MigrationValidationError(
                f"unexpected tables after upgrade: {sorted(tables)}"
            )

        validate_constraints(root, container_name)
        print("Upgrade and constraint validation: PASSED")

        apply_migrations(root, container_name, down_files)

        remaining_tables = current_tables(root, container_name)
        if remaining_tables:
            raise MigrationValidationError(
                "tables remain after rollback: "
                f"{sorted(remaining_tables)}"
            )

        print("Rollback validation: PASSED")
        print("MIGRATION VALIDATION PASSED")
        return 0

    finally:
        subprocess.run(
            ["docker", "rm", "--force", container_name],
            cwd=root,
            text=True,
            capture_output=False,
            check=False,
        )


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except MigrationValidationError as error:
        print(f"ERROR: {error}", file=sys.stderr)
        raise SystemExit(1) from error
