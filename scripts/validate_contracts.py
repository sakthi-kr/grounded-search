#!/usr/bin/env python3
"""Validate GroundedSearch Phase 1 OpenAPI contracts."""

from __future__ import annotations

import sys
from collections.abc import Mapping
from pathlib import Path
from typing import Any

import yaml

REPOSITORY_ROOT = Path(__file__).resolve().parents[1]

CONTRACTS: dict[Path, dict[str, set[str]]] = {
    REPOSITORY_ROOT / "api" / "internal" / "ml-service-v1.yaml": {
        "paths": {"/healthz", "/readyz", "/v1/model/info"},
        "schemas": {
            "HealthResponse",
            "ReadinessResponse",
            "ModelInfoResponse",
            "ErrorResponse",
            "ErrorDetail",
        },
    },
    REPOSITORY_ROOT / "api" / "openapi" / "groundedsearch-v1.yaml": {
        "paths": {"/healthz", "/readyz", "/v1/system/status"},
        "schemas": {
            "HealthResponse",
            "ReadinessResponse",
            "SystemStatusResponse",
            "MLServiceStatus",
        },
    },
}

EXPECTED_PROPERTIES: dict[tuple[str, str], set[str]] = {
    (
        "ml-service-v1.yaml",
        "ModelInfoResponse",
    ): {
        "service",
        "version",
        "status",
        "mode",
        "embedding_model",
        "reranker_model",
        "python_version",
        "environment",
    },
    (
        "groundedsearch-v1.yaml",
        "SystemStatusResponse",
    ): {
        "service",
        "status",
        "version",
        "uptime_seconds",
        "dependencies",
    },
    (
        "groundedsearch-v1.yaml",
        "MLServiceStatus",
    ): {
        "status",
        "version",
        "mode",
        "embedding_model",
        "reranker_model",
        "python_version",
        "environment",
        "error",
    },
}


class ContractError(RuntimeError):
    """Raised when a contract violates the Phase 1 rules."""


def load_contract(path: Path) -> dict[str, Any]:
    text = path.read_text(encoding="utf-8")
    if "\t" in text:
        raise ContractError(f"{path}: tab characters are not allowed")

    loaded = yaml.safe_load(text)
    if not isinstance(loaded, dict):
        raise ContractError(f"{path}: document root must be an object")

    return loaded


def require_mapping(
    value: Any,
    *,
    path: Path,
    location: str,
) -> Mapping[str, Any]:
    if not isinstance(value, Mapping):
        raise ContractError(f"{path}: {location} must be an object")
    return value


def validate_refs(
    value: Any,
    *,
    path: Path,
    schemas: Mapping[str, Any],
) -> None:
    if isinstance(value, Mapping):
        reference = value.get("$ref")
        if isinstance(reference, str):
            prefix = "#/components/schemas/"
            header_prefix = "#/components/headers/"
            response_prefix = "#/components/responses/"

            if reference.startswith(prefix):
                name = reference.removeprefix(prefix)
                if name not in schemas:
                    raise ContractError(
                        f"{path}: unresolved schema reference {reference}"
                    )
            elif not (
                reference.startswith(header_prefix)
                or reference.startswith(response_prefix)
            ):
                raise ContractError(
                    f"{path}: only local component references are allowed: {reference}"
                )

        for child in value.values():
            validate_refs(child, path=path, schemas=schemas)
        return

    if isinstance(value, list):
        for child in value:
            validate_refs(child, path=path, schemas=schemas)


def validate_contract(
    path: Path,
    expectations: dict[str, set[str]],
) -> None:
    document = load_contract(path)

    if document.get("openapi") != "3.1.0":
        raise ContractError(f"{path}: openapi must equal 3.1.0")

    info = require_mapping(
        document.get("info"),
        path=path,
        location="info",
    )
    if not info.get("title") or not info.get("version"):
        raise ContractError(f"{path}: info.title and info.version are required")

    paths = require_mapping(
        document.get("paths"),
        path=path,
        location="paths",
    )
    actual_paths = set(paths)
    if actual_paths != expectations["paths"]:
        raise ContractError(f"{path}: unexpected path set: {sorted(actual_paths)}")

    operation_ids: set[str] = set()
    for route, route_item in paths.items():
        route_mapping = require_mapping(
            route_item,
            path=path,
            location=f"paths.{route}",
        )
        get_operation = require_mapping(
            route_mapping.get("get"),
            path=path,
            location=f"paths.{route}.get",
        )
        operation_id = get_operation.get("operationId")
        if not isinstance(operation_id, str) or not operation_id:
            raise ContractError(f"{path}: {route} is missing operationId")
        if operation_id in operation_ids:
            raise ContractError(f"{path}: duplicate operationId {operation_id}")
        operation_ids.add(operation_id)

        responses = require_mapping(
            get_operation.get("responses"),
            path=path,
            location=f"paths.{route}.get.responses",
        )
        if "200" not in responses:
            raise ContractError(f"{path}: {route} is missing a 200 response")

    components = require_mapping(
        document.get("components"),
        path=path,
        location="components",
    )
    schemas = require_mapping(
        components.get("schemas"),
        path=path,
        location="components.schemas",
    )
    actual_schemas = set(schemas)
    missing_schemas = expectations["schemas"] - actual_schemas
    if missing_schemas:
        raise ContractError(f"{path}: missing schemas: {sorted(missing_schemas)}")

    for schema_name, schema_value in schemas.items():
        schema = require_mapping(
            schema_value,
            path=path,
            location=f"components.schemas.{schema_name}",
        )
        if schema.get("type") == "object":
            if schema.get("additionalProperties") is not False:
                raise ContractError(
                    f"{path}: {schema_name} must forbid additional properties"
                )

    for (filename, schema_name), expected in EXPECTED_PROPERTIES.items():
        if path.name != filename:
            continue

        schema = require_mapping(
            schemas.get(schema_name),
            path=path,
            location=f"components.schemas.{schema_name}",
        )
        properties = require_mapping(
            schema.get("properties"),
            path=path,
            location=f"components.schemas.{schema_name}.properties",
        )
        if set(properties) != expected:
            raise ContractError(
                f"{path}: {schema_name} properties differ from the versioned contract"
            )

    validate_refs(document, path=path, schemas=schemas)


def main() -> int:
    for contract_path, expectations in CONTRACTS.items():
        validate_contract(contract_path, expectations)
        print(f"validated: {contract_path.relative_to(REPOSITORY_ROOT)}")

    print(f"contract validation passed: {len(CONTRACTS)} files")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (ContractError, OSError, yaml.YAMLError) as error:
        print(f"contract validation failed: {error}", file=sys.stderr)
        raise SystemExit(1) from error
