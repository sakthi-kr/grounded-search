package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	handler := testHandler()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusOK,
		)
	}

	var payload healthResponse
	decodeJSON(t, response.Body.Bytes(), &payload)

	if payload.Service != serviceName {
		t.Errorf(
			"service = %q, want %q",
			payload.Service,
			serviceName,
		)
	}
	if payload.Status != "healthy" {
		t.Errorf(
			"status = %q, want healthy",
			payload.Status,
		)
	}
	if payload.Version != "test-version" {
		t.Errorf(
			"version = %q, want test-version",
			payload.Version,
		)
	}

	assertCommonHeaders(t, response)
}

func TestReadyEndpoint(t *testing.T) {
	t.Parallel()

	handler := testHandler()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusOK,
		)
	}

	var payload healthResponse
	decodeJSON(t, response.Body.Bytes(), &payload)

	if payload.Status != "ready" {
		t.Errorf("status = %q, want ready", payload.Status)
	}

	assertCommonHeaders(t, response)
}

func TestSystemStatusEndpoint(t *testing.T) {
	t.Parallel()

	handler := NewHandler(
		discardLogger(),
		"test-version",
		time.Now().Add(-5*time.Second),
	)
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/system/status",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusOK,
		)
	}

	var payload systemStatusResponse
	decodeJSON(t, response.Body.Bytes(), &payload)

	if payload.Status != "foundation" {
		t.Errorf(
			"status = %q, want foundation",
			payload.Status,
		)
	}
	if payload.Dependencies["ml_service"] != "not_configured" {
		t.Errorf(
			"ml_service = %q, want not_configured",
			payload.Dependencies["ml_service"],
		)
	}
	if payload.UptimeSeconds < 4 {
		t.Errorf(
			"uptime_seconds = %d, want at least 4",
			payload.UptimeSeconds,
		)
	}

	assertCommonHeaders(t, response)
}

func TestUnknownRouteReturnsStructuredNotFound(t *testing.T) {
	t.Parallel()

	handler := testHandler()
	request := httptest.NewRequest(
		http.MethodGet,
		"/does-not-exist",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusNotFound,
		)
	}

	var payload apiErrorBody
	decodeJSON(t, response.Body.Bytes(), &payload)

	if payload.Error.Code != "not_found" {
		t.Errorf(
			"error code = %q, want not_found",
			payload.Error.Code,
		)
	}
	if payload.Error.RequestID == "" {
		t.Error("request_id is empty")
	}
}

func TestUnsupportedMethodReturnsStructuredError(t *testing.T) {
	t.Parallel()

	handler := testHandler()
	request := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusMethodNotAllowed,
		)
	}
	if response.Header().Get("Allow") != http.MethodGet {
		t.Errorf(
			"Allow = %q, want GET",
			response.Header().Get("Allow"),
		)
	}

	var payload apiErrorBody
	decodeJSON(t, response.Body.Bytes(), &payload)

	if payload.Error.Code != "method_not_allowed" {
		t.Errorf(
			"error code = %q, want method_not_allowed",
			payload.Error.Code,
		)
	}
}

func TestRequestIDIsPreserved(t *testing.T) {
	t.Parallel()

	handler := testHandler()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(requestIDHeader, "client-request-123")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Header().Get(requestIDHeader) != "client-request-123" {
		t.Errorf(
			"X-Request-ID = %q, want client-request-123",
			response.Header().Get(requestIDHeader),
		)
	}
}

func TestInvalidRequestIDIsReplaced(t *testing.T) {
	t.Parallel()

	handler := testHandler()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(requestIDHeader, "invalid request id")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	requestID := response.Header().Get(requestIDHeader)
	if requestID == "" {
		t.Fatal("X-Request-ID is empty")
	}
	if requestID == "invalid request id" {
		t.Error("invalid client request ID was not replaced")
	}
}

func TestRecoveryMiddlewareReturnsStructuredError(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	panicHandler := http.HandlerFunc(func(
		http.ResponseWriter,
		*http.Request,
	) {
		panic("test panic")
	})

	handler := requestIDMiddleware(
		recoveryMiddleware(logger, panicHandler),
	)

	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusInternalServerError,
		)
	}

	var payload apiErrorBody
	decodeJSON(t, response.Body.Bytes(), &payload)

	if payload.Error.Code != "internal_error" {
		t.Errorf(
			"error code = %q, want internal_error",
			payload.Error.Code,
		)
	}
	if !strings.Contains(logs.String(), "panic recovered") {
		t.Error("panic was not logged")
	}
}

func testHandler() http.Handler {
	return NewHandler(
		discardLogger(),
		"test-version",
		time.Now(),
	)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func decodeJSON(
	t *testing.T,
	data []byte,
	target any,
) {
	t.Helper()

	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode JSON: %v\nbody: %s", err, data)
	}
}

func assertCommonHeaders(
	t *testing.T,
	response *httptest.ResponseRecorder,
) {
	t.Helper()

	if response.Header().Get("Content-Type") !=
		"application/json; charset=utf-8" {
		t.Errorf(
			"Content-Type = %q",
			response.Header().Get("Content-Type"),
		)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Errorf(
			"Cache-Control = %q, want no-store",
			response.Header().Get("Cache-Control"),
		)
	}
	if response.Header().Get(requestIDHeader) == "" {
		t.Error("X-Request-ID is empty")
	}
}
