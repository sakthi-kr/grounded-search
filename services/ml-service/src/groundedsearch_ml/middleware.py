"""HTTP request context, access logging, and safe exception handling."""

from __future__ import annotations

import logging
from time import perf_counter
from typing import Any
from uuid import uuid4

from starlette.datastructures import Headers, MutableHeaders
from starlette.responses import JSONResponse
from starlette.types import ASGIApp, Message, Receive, Scope, Send

from groundedsearch_ml.schemas import ErrorDetail, ErrorResponse

_REQUEST_ID_HEADER = "X-Request-ID"


class RequestContextMiddleware:
    """Attach request IDs, security headers, and structured access logs."""

    def __init__(
        self,
        app: ASGIApp,
        logger: logging.Logger,
    ) -> None:
        self._app = app
        self._logger = logger

    async def __call__(
        self,
        scope: Scope,
        receive: Receive,
        send: Send,
    ) -> None:
        if scope["type"] != "http":
            await self._app(scope, receive, send)
            return

        headers = Headers(scope=scope)
        request_id = _normalise_request_id(headers.get(_REQUEST_ID_HEADER))
        if request_id is None:
            request_id = uuid4().hex

        state = scope.setdefault("state", {})
        state["request_id"] = request_id

        started = perf_counter()
        status_code = 500
        response_started = False

        async def send_wrapper(message: Message) -> None:
            nonlocal response_started, status_code

            if message["type"] == "http.response.start":
                response_started = True
                status_code = int(message["status"])
                response_headers = MutableHeaders(scope=message)
                response_headers[_REQUEST_ID_HEADER] = request_id
                response_headers["Cache-Control"] = "no-store"

            await send(message)

        try:
            await self._app(scope, receive, send_wrapper)
        except Exception:
            self._logger.exception(
                "unhandled request exception",
                extra={
                    "request_id": request_id,
                    "method": scope.get("method"),
                    "path": scope.get("path"),
                },
            )

            if response_started:
                raise

            error_response = ErrorResponse(
                error=ErrorDetail(
                    code="internal_error",
                    message="an internal error occurred",
                    request_id=request_id,
                )
            )
            response = JSONResponse(
                status_code=500,
                content=error_response.model_dump(),
            )
            await response(scope, receive, send_wrapper)
        finally:
            self._logger.info(
                "HTTP request completed",
                extra={
                    "request_id": request_id,
                    "method": scope.get("method"),
                    "path": scope.get("path"),
                    "status": status_code,
                    "duration_ms": round(
                        (perf_counter() - started) * 1000,
                        3,
                    ),
                },
            )


def request_id_from_scope(scope: Scope) -> str | None:
    """Return the request ID previously stored by the middleware."""

    state: dict[str, Any] = scope.get("state", {})
    value = state.get("request_id")
    return value if isinstance(value, str) else None


def _normalise_request_id(raw: str | None) -> str | None:
    if raw is None:
        return None

    value = raw.strip()
    if not value or len(value) > 128:
        return None

    if any(ord(character) < 0x21 or ord(character) > 0x7E for character in value):
        return None

    return value
