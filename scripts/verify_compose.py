#!/usr/bin/env python3
"""Build and smoke-test the Phase 1 Docker Compose environment."""

from __future__ import annotations

import json
import os
from pathlib import Path
import socket
import subprocess
import sys
import time
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

COMPOSE_FILE = Path("deploy/docker-compose/compose.yaml")
HEALTH_TIMEOUT_SECONDS = 90.0
REQUEST_TIMEOUT_SECONDS = 3.0


class ComposeVerificationError(RuntimeError):
    """Raised when the Compose smoke test fails."""


def run(
    command: list[str],
    *,
    cwd: Path,
    environment: dict[str, str],
    capture: bool = True,
) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(
        command,
        cwd=cwd,
        env=environment,
        text=True,
        capture_output=capture,
        check=False,
    )
    if result.returncode != 0:
        raise ComposeVerificationError(
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
        raise ComposeVerificationError(
            "run this script from inside the GroundedSearch repository"
        )
    return Path(result.stdout.strip()).resolve()


def free_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as listener:
        listener.bind(("127.0.0.1", 0))
        return int(listener.getsockname()[1])


def request_json(url: str, request_id: str) -> tuple[int, dict[str, Any], str]:
    request = Request(url, headers={"X-Request-ID": request_id})

    try:
        with urlopen(request, timeout=REQUEST_TIMEOUT_SECONDS) as response:
            payload = json.loads(response.read().decode("utf-8"))
            response_request_id = response.headers.get("X-Request-ID", "")
            return response.status, payload, response_request_id
    except HTTPError as error:
        body = error.read().decode("utf-8", errors="replace")
        raise ComposeVerificationError(
            f"{url} returned HTTP {error.code}: {body}"
        ) from error
    except (URLError, TimeoutError, json.JSONDecodeError) as error:
        raise ComposeVerificationError(
            f"request failed for {url}: {error}"
        ) from error


def wait_for_status(
    url: str,
    expected_status: str,
    *,
    deadline: float,
) -> dict[str, Any]:
    last_error = "no request attempted"

    while time.monotonic() < deadline:
        try:
            status_code, payload, _ = request_json(
                url,
                "compose-readiness-check",
            )
            if (
                status_code == 200
                and payload.get("status") == expected_status
            ):
                return payload
            last_error = (
                f"status_code={status_code}, payload_status="
                f"{payload.get('status')!r}"
            )
        except ComposeVerificationError as error:
            last_error = str(error)

        time.sleep(1.0)

    raise ComposeVerificationError(
        f"timed out waiting for {url} to report "
        f"{expected_status!r}: {last_error}"
    )


def verify_healthy(search_port: int) -> None:
    request_id = "compose-healthy-test"
    status_code, payload, response_request_id = request_json(
        f"http://127.0.0.1:{search_port}/v1/system/status",
        request_id,
    )

    if status_code != 200:
        raise ComposeVerificationError(
            f"healthy status code was {status_code}, expected 200"
        )
    if response_request_id != request_id:
        raise ComposeVerificationError(
            "search API did not preserve the request ID"
        )
    if payload.get("status") != "healthy":
        raise ComposeVerificationError(
            f"search API status was {payload.get('status')!r}, "
            "expected 'healthy'"
        )

    ml_service = payload.get("dependencies", {}).get("ml_service", {})
    if ml_service.get("status") != "ready":
        raise ComposeVerificationError(
            f"ML dependency status was {ml_service.get('status')!r}, "
            "expected 'ready'"
        )
    if ml_service.get("version") != "0.1.0":
        raise ComposeVerificationError(
            f"ML service version was {ml_service.get('version')!r}, "
            "expected '0.1.0'"
        )


def verify_degraded(search_port: int) -> None:
    payload = wait_for_status(
        f"http://127.0.0.1:{search_port}/v1/system/status",
        "degraded",
        deadline=time.monotonic() + 30.0,
    )

    ml_service = payload.get("dependencies", {}).get("ml_service", {})
    actual = (
        ml_service.get("status"),
        ml_service.get("error"),
    )
    allowed = {
        ("unavailable", "dependency_unavailable"),
        ("timeout", "dependency_timeout"),
    }

    if actual not in allowed:
        raise ComposeVerificationError(
            "unexpected degraded ML dependency state: "
            f"status={actual[0]!r}, error={actual[1]!r}; "
            f"expected one of {sorted(allowed)!r}"
        )


def main() -> int:
    root = repository_root()
    compose_path = root / COMPOSE_FILE
    if not compose_path.is_file():
        raise ComposeVerificationError(
            f"Compose file is missing: {compose_path}"
        )

    search_port = free_port()
    ml_port = free_port()
    project_name = f"grounded-search-smoke-{os.getpid()}"

    environment = os.environ.copy()
    environment.update(
        {
            "SEARCH_API_HOST_PORT": str(search_port),
            "ML_SERVICE_HOST_PORT": str(ml_port),
            "COMPOSE_PROJECT_NAME": project_name,
        }
    )

    compose = [
        "docker",
        "compose",
        "-f",
        str(compose_path),
    ]

    print(f"Compose project: {project_name}")
    print(f"Search API host port: {search_port}")
    print(f"ML service host port: {ml_port}")

    try:
        run(
            compose + ["config", "--quiet"],
            cwd=root,
            environment=environment,
            capture=False,
        )
        run(
            compose + ["build", "--pull"],
            cwd=root,
            environment=environment,
            capture=False,
        )
        run(
            compose + ["up", "-d", "--wait"],
            cwd=root,
            environment=environment,
            capture=False,
        )

        deadline = time.monotonic() + HEALTH_TIMEOUT_SECONDS
        wait_for_status(
            f"http://127.0.0.1:{ml_port}/readyz",
            "ready",
            deadline=deadline,
        )
        wait_for_status(
            f"http://127.0.0.1:{search_port}/v1/system/status",
            "healthy",
            deadline=deadline,
        )
        verify_healthy(search_port)
        print("Healthy service-to-service status: PASSED")

        run(
            compose + ["stop", "ml-service"],
            cwd=root,
            environment=environment,
            capture=False,
        )
        verify_degraded(search_port)
        print("Degraded search API status: PASSED")

    finally:
        subprocess.run(
            compose
            + [
                "down",
                "--volumes",
                "--remove-orphans",
                "--timeout",
                "10",
            ],
            cwd=root,
            env=environment,
            text=True,
            capture_output=False,
            check=False,
        )

    print("DOCKER COMPOSE VERIFICATION PASSED")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except ComposeVerificationError as error:
        print(f"ERROR: {error}", file=sys.stderr)
        raise SystemExit(1) from error
