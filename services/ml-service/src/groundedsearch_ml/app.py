"""FastAPI application for the GroundedSearch ML service."""

from __future__ import annotations

import logging
import platform
from typing import cast

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException as StarletteHTTPException

from groundedsearch_ml import __version__
from groundedsearch_ml.config import Settings
from groundedsearch_ml.logging_config import configure_logging
from groundedsearch_ml.middleware import (
    RequestContextMiddleware,
    request_id_from_scope,
)
from groundedsearch_ml.schemas import (
    ErrorDetail,
    ErrorResponse,
    HealthResponse,
    ModelInfoResponse,
)


def create_app(
    settings: Settings | None = None,
) -> FastAPI:
    """Create a configured ML service application."""

    resolved_settings = settings or Settings.from_env()
    configure_logging(resolved_settings.numeric_log_level)

    logger = logging.getLogger("groundedsearch_ml.http")

    app = FastAPI(
        title="GroundedSearch ML Service",
        version=__version__,
        docs_url="/docs",
        redoc_url=None,
        openapi_url="/openapi.json",
    )
    app.state.settings = resolved_settings
    app.add_middleware(
        RequestContextMiddleware,
        logger=logger,
    )

    @app.exception_handler(StarletteHTTPException)
    async def http_exception_handler(
        request: Request,
        exception: StarletteHTTPException,
    ) -> JSONResponse:
        code = {
            404: "not_found",
            405: "method_not_allowed",
        }.get(exception.status_code, "http_error")

        message = {
            404: "the requested resource was not found",
            405: "the requested method is not allowed",
        }.get(exception.status_code, str(exception.detail))

        payload = ErrorResponse(
            error=ErrorDetail(
                code=code,
                message=message,
                request_id=request_id_from_scope(request.scope),
            )
        )

        return JSONResponse(
            status_code=exception.status_code,
            content=payload.model_dump(),
            headers=exception.headers,
        )

    @app.exception_handler(RequestValidationError)
    async def validation_exception_handler(
        request: Request,
        exception: RequestValidationError,
    ) -> JSONResponse:
        del exception

        payload = ErrorResponse(
            error=ErrorDetail(
                code="validation_error",
                message="the request did not match the API contract",
                request_id=request_id_from_scope(request.scope),
            )
        )

        return JSONResponse(
            status_code=422,
            content=payload.model_dump(),
        )

    @app.get(
        "/healthz",
        response_model=HealthResponse,
        tags=["system"],
    )
    async def health() -> HealthResponse:
        return HealthResponse(
            status="healthy",
            version=__version__,
        )

    @app.get(
        "/readyz",
        response_model=HealthResponse,
        tags=["system"],
    )
    async def readiness() -> HealthResponse:
        return HealthResponse(
            status="ready",
            version=__version__,
        )

    @app.get(
        "/v1/model/info",
        response_model=ModelInfoResponse,
        tags=["models"],
    )
    async def model_info(
        request: Request,
    ) -> ModelInfoResponse:
        current_settings = cast(
            Settings,
            request.app.state.settings,
        )

        return ModelInfoResponse(
            version=__version__,
            python_version=platform.python_version(),
            environment=current_settings.environment,
        )

    return app
