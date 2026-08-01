package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/mlclient"
)

type fakeModelInfoClient struct {
	info              mlclient.ModelInfo
	err               error
	receivedRequestID string
}

func (client *fakeModelInfoClient) ModelInfo(
	ctx context.Context,
	requestID string,
) (mlclient.ModelInfo, error) {
	client.receivedRequestID = requestID
	return client.info, client.err
}

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	handler := testHandler(readyModelInfoClient())
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

	handler := testHandler(readyModelInfoClient())
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

func TestSystemStatusEndpointHealthy(t *testing.T) {
	t.Parallel()

	client := readyModelInfoClient()
	handler := NewHandler(
		discardLogger(),
		"test-version",
		time.Now().Add(-5*time.Second),
		client,
	)
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/system/status",
		nil,
	)
	request.Header.Set(requestIDHeader, "status-request-123")
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

	if payload.Status != "healthy" {
		t.Errorf(
			"status = %q, want healthy",
			payload.Status,
		)
	}
	if payload.Dependencies.MLService.Status != "ready" {
		t.Errorf(
			"ML status = %q, want ready",
			payload.Dependencies.MLService.Status,
		)
	}
	if payload.Dependencies.MLService.Version == nil ||
		*payload.Dependencies.MLService.Version != "0.1.0" {
		t.Errorf(
			"ML version = %v, want 0.1.0",
			payload.Dependencies.MLService.Version,
		)
	}
	if payload.Dependencies.MLService.Error != "" {
		t.Errorf(
			"ML error = %q, want empty",
			payload.Dependencies.MLService.Error,
		)
	}
	if payload.UptimeSeconds < 4 {
		t.Errorf(
			"uptime_seconds = %d, want at least 4",
			payload.UptimeSeconds,
		)
	}
	if client.receivedRequestID != "status-request-123" {
		t.Errorf(
			"forwarded request ID = %q, want status-request-123",
			client.receivedRequestID,
		)
	}

	assertCommonHeaders(t, response)
}

func TestSystemStatusEndpointDegraded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus string
		wantCode   string
	}{
		{
			name:       "timeout",
			err:        context.DeadlineExceeded,
			wantStatus: "timeout",
			wantCode:   "dependency_timeout",
		},
		{
			name: "invalid response",
			err: errors.Join(
				mlclient.ErrInvalidResponse,
				errors.New("bad payload"),
			),
			wantStatus: "invalid_response",
			wantCode:   "invalid_dependency_response",
		},
		{
			name:       "unavailable",
			err:        errors.New("connection refused"),
			wantStatus: "unavailable",
			wantCode:   "dependency_unavailable",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeModelInfoClient{err: test.err}
			handler := testHandler(client)
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

			if payload.Status != "degraded" {
				t.Errorf(
					"system status = %q, want degraded",
					payload.Status,
				)
			}

			dependency := payload.Dependencies.MLService
			if dependency.Status != test.wantStatus {
				t.Errorf(
					"dependency status = %q, want %q",
					dependency.Status,
					test.wantStatus,
				)
			}
			if dependency.Error != test.wantCode {
				t.Errorf(
					"dependency error = %q, want %q",
					dependency.Error,
					test.wantCode,
				)
			}
			if dependency.Version != nil {
				t.Errorf(
					"dependency version = %v, want nil",
					dependency.Version,
				)
			}
		})
	}
}

func TestSystemStatusWithoutClientIsDegraded(t *testing.T) {
	t.Parallel()

	handler := testHandler(nil)
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/system/status",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	var payload systemStatusResponse
	decodeJSON(t, response.Body.Bytes(), &payload)

	if payload.Status != "degraded" {
		t.Errorf(
			"status = %q, want degraded",
			payload.Status,
		)
	}
	if payload.Dependencies.MLService.Status != "unavailable" {
		t.Errorf(
			"ML status = %q, want unavailable",
			payload.Dependencies.MLService.Status,
		)
	}
}

func TestUnknownRouteReturnsStructuredNotFound(t *testing.T) {
	t.Parallel()

	handler := testHandler(readyModelInfoClient())
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

	handler := testHandler(readyModelInfoClient())
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

	handler := testHandler(readyModelInfoClient())
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

	handler := testHandler(readyModelInfoClient())
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
	if !bytes.Contains(logs.Bytes(), []byte("panic recovered")) {
		t.Error("panic was not logged")
	}
}

func testHandler(client modelInfoProvider) http.Handler {
	return NewHandler(
		discardLogger(),
		"test-version",
		time.Now(),
		client,
	)
}

func readyModelInfoClient() *fakeModelInfoClient {
	return &fakeModelInfoClient{
		info: mlclient.ModelInfo{
			Service:       "ml-service",
			Version:       "0.1.0",
			Status:        "ready",
			Mode:          "foundation",
			PythonVersion: "3.10.5",
			Environment:   "test",
		},
	}
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
