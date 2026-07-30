"""Response schemas exposed by the ML service."""

from typing import Literal

from pydantic import BaseModel, ConfigDict


class StrictResponseModel(BaseModel):
    """Base model that rejects unplanned response fields."""

    model_config = ConfigDict(extra="forbid")


class HealthResponse(StrictResponseModel):
    """Health or readiness response."""

    service: Literal["ml-service"] = "ml-service"
    status: Literal["healthy", "ready"]
    version: str


class ModelInfoResponse(StrictResponseModel):
    """Current model and service metadata."""

    service: Literal["ml-service"] = "ml-service"
    status: Literal["ready"] = "ready"
    mode: Literal["foundation"] = "foundation"
    embedding_model: str | None = None
    reranker_model: str | None = None
    python_version: str
    environment: str


class ErrorDetail(StrictResponseModel):
    """Stable public error details."""

    code: str
    message: str
    request_id: str | None = None


class ErrorResponse(StrictResponseModel):
    """Stable public error envelope."""

    error: ErrorDetail
