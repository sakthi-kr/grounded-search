package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

const startupTimeout = 30 * time.Second

type systemStatus struct {
	Service      string                    `json:"service"`
	Status       string                    `json:"status"`
	Version      string                    `json:"version"`
	Dependencies map[string]dependencyInfo `json:"dependencies"`
}

type dependencyInfo struct {
	Status      string  `json:"status"`
	Version     *string `json:"version"`
	Mode        *string `json:"mode"`
	Environment *string `json:"environment"`
	Error       *string `json:"error"`
}

type managedProcess struct {
	command *exec.Cmd
	output  *bytes.Buffer
}

func TestSearchAPIReportsHealthyAndDegradedMLService(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped with -short")
	}

	repositoryRoot := findRepositoryRoot(t)
	python := selectPython(t, repositoryRoot)
	mlPort := reservePort(t)
	apiPort := reservePort(t)

	mlProcess := startProcess(
		t,
		python,
		[]string{"-m", "groundedsearch_ml"},
		filepath.Join(repositoryRoot, "services", "ml-service"),
		setEnvironment(
			os.Environ(),
			map[string]string{
				"PYTHONPATH": filepath.Join(
					repositoryRoot,
					"services",
					"ml-service",
					"src",
				),
				"ML_SERVICE_HOST": "127.0.0.1",
				"ML_SERVICE_PORT": strconv.Itoa(mlPort),
				"LOG_LEVEL":       "ERROR",
			},
		),
	)
	defer mlProcess.stop(t)

	mlBaseURL := fmt.Sprintf("http://127.0.0.1:%d", mlPort)
	waitForHTTP(t, mlBaseURL+"/healthz", http.StatusOK, mlProcess)

	binaryPath := buildSearchAPI(t, repositoryRoot)
	apiProcess := startProcess(
		t,
		binaryPath,
		nil,
		filepath.Join(repositoryRoot, "apps", "search-api"),
		setEnvironment(
			os.Environ(),
			map[string]string{
				"SEARCH_API_HOST":    "127.0.0.1",
				"SEARCH_API_PORT":    strconv.Itoa(apiPort),
				"ML_SERVICE_URL":     mlBaseURL,
				"ML_SERVICE_TIMEOUT": "500ms",
				"LOG_LEVEL":          "ERROR",
			},
		),
	)
	defer apiProcess.stop(t)

	apiBaseURL := fmt.Sprintf("http://127.0.0.1:%d", apiPort)
	waitForHTTP(t, apiBaseURL+"/healthz", http.StatusOK, apiProcess)

	requestID := "integration-request-id"
	response, status := getSystemStatus(t, apiBaseURL, requestID)

	if status != http.StatusOK {
		t.Fatalf("status code = %d, want 200", status)
	}
	if response.Service != "search-api" {
		t.Errorf("service = %q, want search-api", response.Service)
	}
	if response.Status != "healthy" {
		t.Errorf("status = %q, want healthy", response.Status)
	}

	mlDependency, exists := response.Dependencies["ml_service"]
	if !exists {
		t.Fatal("ml_service dependency is missing")
	}
	if mlDependency.Status != "ready" {
		t.Errorf("ml_service status = %q, want ready", mlDependency.Status)
	}
	if value(mlDependency.Version) != "0.1.0" {
		t.Errorf("ml_service version = %q, want 0.1.0", value(mlDependency.Version))
	}
	if value(mlDependency.Mode) != "foundation" {
		t.Errorf("ml_service mode = %q, want foundation", value(mlDependency.Mode))
	}

	mlProcess.stop(t)

	deadline := time.Now().Add(10 * time.Second)
	for {
		response, status = getSystemStatus(t, apiBaseURL, "degraded-request")
		dependency := response.Dependencies["ml_service"]
		if status == http.StatusOK &&
			response.Status == "degraded" &&
			dependency.Status == "unavailable" &&
			value(dependency.Error) == "dependency_unavailable" {
			break
		}

		if time.Now().After(deadline) {
			t.Fatalf(
				"degraded state not observed; response=%+v status=%d",
				response,
				status,
			)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func findRepositoryRoot(t *testing.T) string {
	t.Helper()

	currentDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	for {
		required := filepath.Join(
			currentDirectory,
			"apps",
			"search-api",
			"go.mod",
		)
		if information, err := os.Stat(required); err == nil && !information.IsDir() {
			return currentDirectory
		}

		parent := filepath.Dir(currentDirectory)
		if parent == currentDirectory {
			t.Fatal("could not locate repository root")
		}
		currentDirectory = parent
	}
}

func selectPython(t *testing.T, repositoryRoot string) string {
	t.Helper()

	if configured := strings.TrimSpace(os.Getenv("GROUNDEDSEARCH_PYTHON")); configured != "" {
		return configured
	}

	candidates := []string{
		filepath.Join(
			repositoryRoot,
			"services",
			"ml-service",
			".venv",
			"Scripts",
			"python.exe",
		),
		filepath.Join(
			repositoryRoot,
			"services",
			"ml-service",
			".venv",
			"bin",
			"python",
		),
	}

	for _, candidate := range candidates {
		if information, err := os.Stat(candidate); err == nil && !information.IsDir() {
			return candidate
		}
	}

	python, err := exec.LookPath("python")
	if err == nil {
		return python
	}

	python3, err := exec.LookPath("python3")
	if err == nil {
		return python3
	}

	t.Fatal("Python executable not found")
	return ""
}

func reservePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer listener.Close()

	address, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected listener address: %T", listener.Addr())
	}
	return address.Port
}

func buildSearchAPI(t *testing.T, repositoryRoot string) string {
	t.Helper()

	binaryName := "search-api"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(t.TempDir(), binaryName)

	command := exec.Command(
		"go",
		"build",
		"-o",
		binaryPath,
		"./cmd/server",
	)
	command.Dir = filepath.Join(repositoryRoot, "apps", "search-api")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build search API: %v\n%s", err, output)
	}

	return binaryPath
}

func startProcess(
	t *testing.T,
	name string,
	arguments []string,
	directory string,
	environment []string,
) *managedProcess {
	t.Helper()

	output := &bytes.Buffer{}
	command := exec.Command(name, arguments...)
	command.Dir = directory
	command.Env = environment
	command.Stdout = output
	command.Stderr = output

	if err := command.Start(); err != nil {
		t.Fatalf("start %s: %v", name, err)
	}

	return &managedProcess{command: command, output: output}
}

func (process *managedProcess) stop(t *testing.T) {
	t.Helper()

	if process == nil || process.command == nil || process.command.Process == nil {
		return
	}
	if process.command.ProcessState != nil && process.command.ProcessState.Exited() {
		return
	}

	_ = process.command.Process.Kill()
	_, _ = process.command.Process.Wait()
}

func waitForHTTP(
	t *testing.T,
	url string,
	expectedStatus int,
	process *managedProcess,
) {
	t.Helper()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(startupTimeout)

	for {
		request, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			url,
			nil,
		)
		if err != nil {
			t.Fatalf("create health request: %v", err)
		}

		response, err := client.Do(request)
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode == expectedStatus {
				return
			}
		}

		if process.command.ProcessState != nil && process.command.ProcessState.Exited() {
			t.Fatalf("process exited before readiness:\n%s", process.output.String())
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s:\n%s", url, process.output.String())
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func getSystemStatus(
	t *testing.T,
	apiBaseURL string,
	requestID string,
) (systemStatus, int) {
	t.Helper()

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		apiBaseURL+"/v1/system/status",
		nil,
	)
	if err != nil {
		t.Fatalf("create status request: %v", err)
	}
	request.Header.Set("X-Request-ID", requestID)

	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("execute status request: %v", err)
	}
	defer response.Body.Close()

	if actual := response.Header.Get("X-Request-ID"); actual != requestID {
		t.Errorf("X-Request-ID = %q, want %q", actual, requestID)
	}

	var payload systemStatus
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("decode status response: %v; remaining body=%q", err, body)
	}

	return payload, response.StatusCode
}

func setEnvironment(environment []string, values map[string]string) []string {
	result := append([]string(nil), environment...)

	for key, value := range values {
		prefix := strings.ToUpper(key) + "="
		filtered := result[:0]
		for _, entry := range result {
			if strings.HasPrefix(strings.ToUpper(entry), prefix) {
				continue
			}
			filtered = append(filtered, entry)
		}
		result = append(filtered, key+"="+value)
	}

	return result
}

func value(pointer *string) string {
	if pointer == nil {
		return ""
	}
	return *pointer
}
