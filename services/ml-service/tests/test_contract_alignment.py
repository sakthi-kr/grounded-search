"""Check that Python response models match the internal contract."""

from pathlib import Path

import yaml

from groundedsearch_ml.schemas import (
    ErrorResponse,
    HealthResponse,
    ModelInfoResponse,
)

REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
CONTRACT_PATH = REPOSITORY_ROOT / "api" / "internal" / "ml-service-v1.yaml"


def test_response_models_match_contract_properties() -> None:
    contract = yaml.safe_load(CONTRACT_PATH.read_text(encoding="utf-8"))
    schemas = contract["components"]["schemas"]

    expected = {
        "HealthResponse": set(HealthResponse.model_fields),
        "ModelInfoResponse": set(ModelInfoResponse.model_fields),
        "ErrorResponse": set(ErrorResponse.model_fields),
    }

    for schema_name, model_fields in expected.items():
        contract_fields = set(schemas[schema_name]["properties"])
        assert contract_fields == model_fields


def test_model_info_required_fields_match_contract() -> None:
    contract = yaml.safe_load(CONTRACT_PATH.read_text(encoding="utf-8"))
    required = set(contract["components"]["schemas"]["ModelInfoResponse"]["required"])

    assert required == set(ModelInfoResponse.model_fields)
