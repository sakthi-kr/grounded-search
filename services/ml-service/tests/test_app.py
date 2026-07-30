"""HTTP tests for the ML service."""

from collections.abc import Iterator

import httpx
import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from groundedsearch_ml import __version__
from groundedsearch_ml.app import create_app
from groundedsearch_ml.config import Settings


@pytest.fixture
def app() -> FastAPI:
    return create_app(
        Settings(
            host="127.0.0.1",
            port=8090,
            log_level="CRITICAL",
            environment="test",
        )
    )


@pytest.fixture
def client(app: FastAPI) -> Iterator[TestClient]:
    with TestClient(
        app,
        raise_server_exceptions=False,
    ) as test_client:
        yield test_client


def test_health(client: TestClient) -> None:
    response = client.get("/healthz")

    assert response.status_code == 200
    assert response.json() == {
        "service": "ml-service",
        "status": "healthy",
        "version": __version__,
    }
    _assert_common_headers(response)


def test_readiness(client: TestClient) -> None:
    response = client.get("/readyz")

    assert response.status_code == 200
    assert response.json() == {
        "service": "ml-service",
        "status": "ready",
        "version": __version__,
    }
    _assert_common_headers(response)


def test_model_info(client: TestClient) -> None:
    response = client.get("/v1/model/info")

    assert response.status_code == 200
    payload = response.json()
    assert payload["service"] == "ml-service"
    assert payload["status"] == "ready"
    assert payload["mode"] == "foundation"
    assert payload["embedding_model"] is None
    assert payload["reranker_model"] is None
    assert payload["environment"] == "test"
    assert payload["python_version"]
    _assert_common_headers(response)


def test_unknown_route(client: TestClient) -> None:
    response = client.get("/missing")

    assert response.status_code == 404
    payload = response.json()
    assert payload["error"]["code"] == "not_found"
    assert payload["error"]["request_id"]


def test_method_not_allowed(client: TestClient) -> None:
    response = client.post("/healthz")

    assert response.status_code == 405
    payload = response.json()
    assert payload["error"]["code"] == "method_not_allowed"
    assert response.headers["allow"] == "GET"


def test_request_id_is_preserved(client: TestClient) -> None:
    response = client.get(
        "/healthz",
        headers={"X-Request-ID": "client-request-123"},
    )

    assert response.headers["X-Request-ID"] == "client-request-123"


def test_invalid_request_id_is_replaced(client: TestClient) -> None:
    response = client.get(
        "/healthz",
        headers={"X-Request-ID": "invalid request id"},
    )

    request_id = response.headers["X-Request-ID"]
    assert request_id
    assert request_id != "invalid request id"


def test_unhandled_exception_returns_structured_error(
    app: FastAPI,
) -> None:
    @app.get("/panic")
    async def panic() -> None:
        raise RuntimeError("test panic")

    with TestClient(
        app,
        raise_server_exceptions=False,
    ) as client:
        response = client.get("/panic")

    assert response.status_code == 500
    payload = response.json()
    assert payload["error"]["code"] == "internal_error"
    assert payload["error"]["request_id"]
    _assert_common_headers(response)


def test_validation_error_is_structured(
    app: FastAPI,
) -> None:
    @app.get("/items/{item_id}")
    async def read_item(item_id: int) -> dict[str, int]:
        return {"item_id": item_id}

    with TestClient(app) as client:
        response = client.get("/items/not-an-integer")

    assert response.status_code == 422
    payload = response.json()
    assert payload["error"]["code"] == "validation_error"
    assert payload["error"]["request_id"]
    _assert_common_headers(response)


def _assert_common_headers(response: httpx.Response) -> None:
    headers = response.headers
    assert headers["content-type"].startswith("application/json")
    assert headers["cache-control"] == "no-store"
    assert headers["x-request-id"]
