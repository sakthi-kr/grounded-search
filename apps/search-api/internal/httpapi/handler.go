package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

const (
	requestIDHeader = "X-Request-ID"
	serviceName     = "search-api"
)

type contextKey string

const requestIDContextKey contextKey = "request-id"

// Handler owns the Phase 1 HTTP routes and middleware.
type Handler struct {
	logger    *slog.Logger
	version   string
	startedAt time.Time
}

// NewHandler returns the complete Phase 1 HTTP handler.
func NewHandler(
	logger *slog.Logger,
	version string,
	startedAt time.Time,
) http.Handler {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	if strings.TrimSpace(version) == "" {
		version = "dev"
	}

	handler := &Handler{
		logger:    logger,
		version:   version,
		startedAt: startedAt.UTC(),
	}

	mux := http.NewServeMux()
	mux.Handle(
		"/healthz",
		requireMethod(
			http.MethodGet,
			http.HandlerFunc(handler.health),
		),
	)
	mux.Handle(
		"/readyz",
		requireMethod(
			http.MethodGet,
			http.HandlerFunc(handler.ready),
		),
	)
	mux.Handle(
		"/v1/system/status",
		requireMethod(
			http.MethodGet,
			http.HandlerFunc(handler.systemStatus),
		),
	)
	mux.Handle("/", http.HandlerFunc(handler.notFound))

	return requestIDMiddleware(
		recoveryMiddleware(
			logger,
			accessLogMiddleware(logger, mux),
		),
	)
}

func (handler *Handler) health(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writeJSON(writer, http.StatusOK, healthResponse{
		Service: serviceName,
		Status:  "healthy",
		Version: handler.version,
	})
}

func (handler *Handler) ready(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writeJSON(writer, http.StatusOK, healthResponse{
		Service: serviceName,
		Status:  "ready",
		Version: handler.version,
	})
}

func (handler *Handler) systemStatus(
	writer http.ResponseWriter,
	request *http.Request,
) {
	uptime := time.Since(handler.startedAt)
	if uptime < 0 {
		uptime = 0
	}

	writeJSON(writer, http.StatusOK, systemStatusResponse{
		Service:       serviceName,
		Status:        "foundation",
		Version:       handler.version,
		UptimeSeconds: int64(uptime / time.Second),
		Dependencies: map[string]string{
			"ml_service": "not_configured",
		},
	})
}

func (handler *Handler) notFound(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writeAPIError(
		writer,
		request,
		http.StatusNotFound,
		"not_found",
		"the requested resource was not found",
	)
}

type healthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

type systemStatusResponse struct {
	Service       string            `json:"service"`
	Status        string            `json:"status"`
	Version       string            `json:"version"`
	UptimeSeconds int64             `json:"uptime_seconds"`
	Dependencies  map[string]string `json:"dependencies"`
}

type apiErrorBody struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func writeAPIError(
	writer http.ResponseWriter,
	request *http.Request,
	status int,
	code string,
	message string,
) {
	writeJSON(writer, status, apiErrorBody{
		Error: apiError{
			Code:      code,
			Message:   message,
			RequestID: requestIDFromRequest(request),
		},
	})
}

func writeJSON(
	writer http.ResponseWriter,
	status int,
	payload any,
) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)

	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(true)

	if err := encoder.Encode(payload); err != nil {
		// The response headers may already be committed. Logging is handled by
		// the surrounding request middleware.
		return
	}
}

func requireMethod(
	method string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != method {
			writer.Header().Set("Allow", method)
			writeAPIError(
				writer,
				request,
				http.StatusMethodNotAllowed,
				"method_not_allowed",
				fmt.Sprintf(
					"method %s is not allowed for this resource",
					request.Method,
				),
			)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func requestIDMiddleware(next http.Handler) http.Handler {
	var fallbackCounter atomic.Uint64

	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		requestID := validRequestID(
			request.Header.Get(requestIDHeader),
		)

		if requestID == "" {
			requestID = generateRequestID(&fallbackCounter)
		}

		writer.Header().Set(requestIDHeader, requestID)

		contextWithID := withRequestID(
			request.Context(),
			requestID,
		)
		next.ServeHTTP(
			writer,
			request.WithContext(contextWithID),
		)
	})
}

func recoveryMiddleware(
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			logger.Error(
				"panic recovered",
				"request_id", requestIDFromRequest(request),
				"method", request.Method,
				"path", request.URL.Path,
				"panic", fmt.Sprint(recovered),
				"stack", string(debug.Stack()),
			)

			writeAPIError(
				writer,
				request,
				http.StatusInternalServerError,
				"internal_error",
				"an internal error occurred",
			)
		}()

		next.ServeHTTP(writer, request)
	})
}

func accessLogMiddleware(
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		started := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: writer,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, request)

		logger.Info(
			"HTTP request completed",
			"request_id", requestIDFromRequest(request),
			"method", request.Method,
			"path", request.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (recorder *statusRecorder) WriteHeader(status int) {
	if recorder.wroteHeader {
		return
	}

	recorder.status = status
	recorder.wroteHeader = true
	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *statusRecorder) Write(data []byte) (int, error) {
	if !recorder.wroteHeader {
		recorder.WriteHeader(http.StatusOK)
	}

	return recorder.ResponseWriter.Write(data)
}

func (recorder *statusRecorder) Unwrap() http.ResponseWriter {
	return recorder.ResponseWriter
}

func validRequestID(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > 128 {
		return ""
	}

	for _, character := range value {
		if character < 0x21 || character > 0x7e {
			return ""
		}
	}

	return value
}

func generateRequestID(counter *atomic.Uint64) string {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err == nil {
		return hex.EncodeToString(randomBytes)
	}

	return fmt.Sprintf(
		"fallback-%d-%d",
		time.Now().UnixNano(),
		counter.Add(1),
	)
}
