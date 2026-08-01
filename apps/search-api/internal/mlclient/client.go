package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	modelInfoPath       = "/v1/model/info"
	maxResponseBodySize = 1 << 20
	requestIDHeader     = "X-Request-ID"
)

var (
	// ErrInvalidResponse indicates that the dependency response did not match
	// the versioned contract.
	ErrInvalidResponse = errors.New("invalid ML service response")

	// ErrUnexpectedStatus indicates that the dependency returned a non-200
	// HTTP status.
	ErrUnexpectedStatus = errors.New("unexpected ML service status")
)

// HTTPDoer is implemented by http.Client and test doubles.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client calls the versioned ML service contract.
type Client struct {
	baseURL    *url.URL
	httpClient HTTPDoer
}

// ModelInfo is the response returned by GET /v1/model/info.
type ModelInfo struct {
	Service        string  `json:"service"`
	Version        string  `json:"version"`
	Status         string  `json:"status"`
	Mode           string  `json:"mode"`
	EmbeddingModel *string `json:"embedding_model"`
	RerankerModel  *string `json:"reranker_model"`
	PythonVersion  string  `json:"python_version"`
	Environment    string  `json:"environment"`
}

// New creates an ML service client with a bounded timeout.
func New(baseURL string, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("ML service timeout must be greater than zero")
	}

	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse ML service URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf(
			"ML service URL scheme must be http or https",
		)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("ML service URL must include a host")
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("ML service URL must not include credentials")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf(
			"ML service URL must not include a query or fragment",
		)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")

	return &Client{
		baseURL: parsed,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// NewWithHTTPClient creates a client using an injected HTTP implementation.
func NewWithHTTPClient(
	baseURL string,
	httpClient HTTPDoer,
) (*Client, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("HTTP client is nil")
	}

	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse ML service URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf(
			"ML service URL scheme must be http or https",
		)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("ML service URL must include a host")
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("ML service URL must not include credentials")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf(
			"ML service URL must not include a query or fragment",
		)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")

	return &Client{
		baseURL:    parsed,
		httpClient: httpClient,
	}, nil
}

// ModelInfo retrieves and validates model metadata.
func (client *Client) ModelInfo(
	ctx context.Context,
	requestID string,
) (ModelInfo, error) {
	endpoint := *client.baseURL
	endpoint.Path += modelInfoPath

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint.String(),
		nil,
	)
	if err != nil {
		return ModelInfo{}, fmt.Errorf("create model info request: %w", err)
	}

	if requestID != "" {
		request.Header.Set(requestIDHeader, requestID)
	}
	request.Header.Set("Accept", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return ModelInfo{}, fmt.Errorf("call ML service: %w", err)
	}
	defer response.Body.Close()

	body, err := readBoundedBody(response.Body)
	if err != nil {
		return ModelInfo{}, err
	}

	if response.StatusCode != http.StatusOK {
		return ModelInfo{}, fmt.Errorf(
			"%w: status %d",
			ErrUnexpectedStatus,
			response.StatusCode,
		)
	}

	var modelInfo ModelInfo
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&modelInfo); err != nil {
		return ModelInfo{}, fmt.Errorf(
			"%w: decode JSON: %v",
			ErrInvalidResponse,
			err,
		)
	}

	if err := ensureSingleJSONValue(decoder); err != nil {
		return ModelInfo{}, err
	}

	if err := validateModelInfo(modelInfo); err != nil {
		return ModelInfo{}, err
	}

	return modelInfo, nil
}

func readBoundedBody(reader io.Reader) ([]byte, error) {
	limited := io.LimitReader(reader, maxResponseBodySize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read ML service response: %w", err)
	}
	if len(body) > maxResponseBodySize {
		return nil, fmt.Errorf(
			"%w: response body exceeds %d bytes",
			ErrInvalidResponse,
			maxResponseBodySize,
		)
	}

	return body, nil
}

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return fmt.Errorf(
			"%w: response contains multiple JSON values",
			ErrInvalidResponse,
		)
	}

	return fmt.Errorf(
		"%w: trailing response data: %v",
		ErrInvalidResponse,
		err,
	)
}

func validateModelInfo(modelInfo ModelInfo) error {
	switch {
	case modelInfo.Service != "ml-service":
		return fmt.Errorf(
			"%w: service must equal ml-service",
			ErrInvalidResponse,
		)
	case strings.TrimSpace(modelInfo.Version) == "":
		return fmt.Errorf(
			"%w: version must not be empty",
			ErrInvalidResponse,
		)
	case modelInfo.Status != "ready":
		return fmt.Errorf(
			"%w: status must equal ready",
			ErrInvalidResponse,
		)
	case modelInfo.Mode != "foundation":
		return fmt.Errorf(
			"%w: mode must equal foundation",
			ErrInvalidResponse,
		)
	case strings.TrimSpace(modelInfo.PythonVersion) == "":
		return fmt.Errorf(
			"%w: python_version must not be empty",
			ErrInvalidResponse,
		)
	case strings.TrimSpace(modelInfo.Environment) == "":
		return fmt.Errorf(
			"%w: environment must not be empty",
			ErrInvalidResponse,
		)
	default:
		return nil
	}
}
