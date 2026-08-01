package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/mlclient"
)

const (
	requestIDHeader = "X-Request-ID"
	serviceName     = "search-api"
)

type contextKey string

const requestIDContextKey contextKey = "request-id"

type modelInfoProvider interface {
	ModelInfo(
		context.Context,
		string,
	) (mlclient.ModelInfo, error)
}

// Handler owns the Phase 1 HTTP routes and middleware.
type Handler struct {
	logger          *slog.Logger
	version         string
	startedAt       time.Time
	modelInfoClient modelInfoProvider
}

// NewHandler returns the complete Phase 1 HTTP handler.
func NewHandler(
	logger *slog.Logger,
	version string,
	startedAt time.Time,
	modelInfoClient modelInfoProvider,
) http.Handler {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	if strings.TrimSpace(version) == "" {
		version = "dev"
	}

	handler := &Handler{
		logger:          logger,
		version:         version,
		startedAt:       startedAt.UTC(),
		modelInfoClient: modelInfoClient,
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

	dependency := handler.mlServiceStatus(request)
	status := "healthy"
	if dependency.Status != "ready" {
		status = "degraded"
	}

	writeJSON(writer, http.StatusOK, systemStatusResponse{
		Service:       serviceName,
		Status:        status,
		Version:       handler.version,
		UptimeSeconds: int64(uptime / time.Second),
		Dependencies: dependenciesResponse{
			MLService: dependency,
		},
	})
}

func (handler *Handler) mlServiceStatus(
	request *http.Request,
) mlServiceStatusResponse {
	if handler.modelInfoClient == nil {
		return unavailableMLServiceStatus(
			"dependency_unavailable",
			"unavailable",
		)
	}

	info, err := handler.modelInfoClient.ModelInfo(
		request.Context(),
		requestIDFromRequest(request),
	)
	if err != nil {
		status, code := classifyDependencyError(err)

		handler.logger.Warn(
			"ML service status check failed",
			"request_id", requestIDFromRequest(request),
			"dependency_status", status,
			"error", err,
		)

		return unavailableMLServiceStatus(code, status)
	}

	return mlServiceStatusResponse{
		Status:         "ready",
		Version:        stringPointer(info.Version),
		Mode:           stringPointer(info.Mode),
		EmbeddingModel: info.EmbeddingModel,
		RerankerModel:  info.RerankerModel,
		PythonVersion:  stringPointer(info.PythonVersion),
		Environment:    stringPointer(info.Environment),
	}
}

func classifyDependencyError(err error) (status string, code string) {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout", "dependency_timeout"
	}

	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return "timeout", "dependency_timeout"
	}

	if errors.Is(err, mlclient.ErrInvalidResponse) {
		return "invalid_response", "invalid_dependency_response"
	}

	return "unavailable", "dependency_unavailable"
}

func unavailableMLServiceStatus(
	code string,
	status string,
) mlServiceStatusResponse {
	return mlServiceStatusResponse{
		Status: status,
		Error:  code,
	}
}

func stringPointer(value string) *string {
	return &value
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
	Service       string               `json:"service"`
	Status        string               `json:"status"`
	Version       string               `json:"version"`
	UptimeSeconds int64                `json:"uptime_seconds"`
	Dependencies  dependenciesResponse `json:"dependencies"`
}

type dependenciesResponse struct {
	MLService mlServiceStatusResponse `json:"ml_service"`
}

type mlServiceStatusResponse struct {
	Status         string  `json:"status"`
	Version        *string `json:"version"`
	Mode           *string `json:"mode"`
	EmbeddingModel *string `json:"embedding_model"`
	RerankerModel  *string `json:"reranker_model"`
	PythonVersion  *string `json:"python_version"`
	Environment    *string `json:"environment"`
	Error          string  `json:"error,omitempty"`
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
