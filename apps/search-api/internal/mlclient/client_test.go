package mlclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestModelInfoSuccess(t *testing.T) {
	t.Parallel()

	var receivedRequestID string
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		receivedRequestID = request.Header.Get(requestIDHeader)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"service":"ml-service",
			"version":"0.1.0",
			"status":"ready",
			"mode":"foundation",
			"embedding_model":null,
			"reranker_model":null,
			"python_version":"3.10.5",
			"environment":"development"
		}`))
	}))
	defer server.Close()

	client, err := New(server.URL, time.Second)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	info, err := client.ModelInfo(
		context.Background(),
		"request-123",
	)
	if err != nil {
		t.Fatalf("ModelInfo() error = %v", err)
	}

	if receivedRequestID != "request-123" {
		t.Errorf(
			"X-Request-ID = %q, want request-123",
			receivedRequestID,
		)
	}
	if info.PythonVersion != "3.10.5" {
		t.Errorf(
			"PythonVersion = %q, want 3.10.5",
			info.PythonVersion,
		)
	}
}

func TestModelInfoRejectsUnexpectedStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		http.Error(writer, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := New(server.URL, time.Second)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.ModelInfo(context.Background(), "")
	if !errors.Is(err, ErrUnexpectedStatus) {
		t.Fatalf(
			"ModelInfo() error = %v, want ErrUnexpectedStatus",
			err,
		)
	}
}

func TestModelInfoRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		_, _ = writer.Write([]byte(`{"service":`))
	}))
	defer server.Close()

	client, err := New(server.URL, time.Second)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.ModelInfo(context.Background(), "")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf(
			"ModelInfo() error = %v, want ErrInvalidResponse",
			err,
		)
	}
}

func TestModelInfoRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		_, _ = writer.Write([]byte(`{
			"service":"ml-service",
			"version":"0.1.0",
			"status":"ready",
			"mode":"foundation",
			"embedding_model":null,
			"reranker_model":null,
			"python_version":"3.10.5",
			"environment":"development",
			"unexpected":true
		}`))
	}))
	defer server.Close()

	client, err := New(server.URL, time.Second)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.ModelInfo(context.Background(), "")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf(
			"ModelInfo() error = %v, want ErrInvalidResponse",
			err,
		)
	}
}

func TestModelInfoTimesOut(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		time.Sleep(100 * time.Millisecond)
		_, _ = writer.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, err := New(server.URL, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.ModelInfo(context.Background(), "")
	if err == nil {
		t.Fatal("ModelInfo() error = nil, want timeout")
	}

	if !strings.Contains(err.Error(), "Client.Timeout") &&
		!errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ModelInfo() error = %v, want timeout", err)
	}
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		timeout time.Duration
	}{
		{
			name:    "missing scheme",
			baseURL: "localhost:8090",
			timeout: time.Second,
		},
		{
			name:    "unsupported scheme",
			baseURL: "ftp://localhost:8090",
			timeout: time.Second,
		},
		{
			name:    "credentials",
			baseURL: "http://user:pass@localhost:8090",
			timeout: time.Second,
		},
		{
			name:    "query",
			baseURL: "http://localhost:8090?debug=true",
			timeout: time.Second,
		},
		{
			name:    "zero timeout",
			baseURL: "http://localhost:8090",
			timeout: 0,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := New(
				test.baseURL,
				test.timeout,
			); err == nil {
				t.Fatal("New() error = nil, want error")
			}
		})
	}
}
