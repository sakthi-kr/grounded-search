package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const usage = `GroundedSearch project commands

Usage:
  go run ./tools/projectctl <command>

Commands:
  help             Show this help text
  test-go          Run Go unit tests with coverage
  test-python      Run Python tests with coverage
  contracts        Validate the versioned API contracts
  build-go         Build the Go search API in a temporary directory
  integration      Run the service integration test suite
  verify-core      Run contracts, unit tests, build, and integration tests
`

type commandRunner interface {
	Run(name string, args []string, directory string, environment []string) error
}

type execRunner struct {
	stdout io.Writer
	stderr io.Writer
}

func (runner execRunner) Run(
	name string,
	args []string,
	directory string,
	environment []string,
) error {
	command := exec.Command(name, args...)
	command.Dir = directory
	command.Env = environment
	command.Stdout = runner.stdout
	command.Stderr = runner.stderr

	fmt.Fprintf(runner.stdout, "$ %s\n", formatCommand(name, args))

	if err := command.Run(); err != nil {
		return fmt.Errorf("run %s: %w", formatCommand(name, args), err)
	}

	return nil
}

type application struct {
	repositoryRoot string
	python         string
	runner         commandRunner
	stdout         io.Writer
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run(arguments []string, stdout io.Writer, stderr io.Writer) error {
	repositoryRoot, err := findRepositoryRoot()
	if err != nil {
		return err
	}

	python, err := selectPython(repositoryRoot)
	if err != nil {
		return err
	}

	app := application{
		repositoryRoot: repositoryRoot,
		python:         python,
		runner: execRunner{
			stdout: stdout,
			stderr: stderr,
		},
		stdout: stdout,
	}

	command := "help"
	if len(arguments) > 0 {
		command = arguments[0]
	}
	if len(arguments) > 1 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(arguments[1:], " "))
	}

	return app.execute(command)
}

func (app application) execute(command string) error {
	switch command {
	case "help", "-h", "--help":
		_, err := io.WriteString(app.stdout, usage)
		return err
	case "test-go":
		return app.testGo()
	case "test-python":
		return app.testPython()
	case "contracts":
		return app.contracts()
	case "build-go":
		return app.buildGo()
	case "integration":
		return app.integration()
	case "verify-core":
		return app.verifyCore()
	default:
		return fmt.Errorf("unknown command %q\n\n%s", command, usage)
	}
}

func (app application) testGo() error {
	return app.runner.Run(
		"go",
		[]string{"test", "./...", "-cover"},
		filepath.Join(app.repositoryRoot, "apps", "search-api"),
		os.Environ(),
	)
}

func (app application) testPython() error {
	return app.runner.Run(
		app.python,
		[]string{"-m", "pytest"},
		filepath.Join(app.repositoryRoot, "services", "ml-service"),
		pythonEnvironment(app.repositoryRoot),
	)
}

func (app application) contracts() error {
	return app.runner.Run(
		app.python,
		[]string{filepath.Join(app.repositoryRoot, "scripts", "validate_contracts.py")},
		app.repositoryRoot,
		pythonEnvironment(app.repositoryRoot),
	)
}

func (app application) buildGo() error {
	buildDirectory, err := os.MkdirTemp("", "groundedsearch-build-")
	if err != nil {
		return fmt.Errorf("create temporary build directory: %w", err)
	}
	defer os.RemoveAll(buildDirectory)

	outputName := "search-api"
	if runtime.GOOS == "windows" {
		outputName += ".exe"
	}

	return app.runner.Run(
		"go",
		[]string{
			"build",
			"-o",
			filepath.Join(buildDirectory, outputName),
			"./cmd/server",
		},
		filepath.Join(app.repositoryRoot, "apps", "search-api"),
		os.Environ(),
	)
}

func (app application) integration() error {
	return app.runner.Run(
		"go",
		[]string{"test", "./...", "-count=1", "-v"},
		filepath.Join(app.repositoryRoot, "tests", "integration"),
		integrationEnvironment(app.repositoryRoot, app.python),
	)
}

func (app application) verifyCore() error {
	steps := []struct {
		name string
		run  func() error
	}{
		{name: "contracts", run: app.contracts},
		{name: "Go unit tests", run: app.testGo},
		{name: "Python unit tests", run: app.testPython},
		{name: "Go build", run: app.buildGo},
		{name: "service integration tests", run: app.integration},
	}

	for _, step := range steps {
		fmt.Fprintf(app.stdout, "\n==> %s\n", step.name)
		if err := step.run(); err != nil {
			return fmt.Errorf("%s failed: %w", step.name, err)
		}
	}

	fmt.Fprintln(app.stdout, "\nCORE VERIFICATION PASSED")
	return nil
}

func findRepositoryRoot() (string, error) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current directory: %w", err)
	}

	for {
		if isRepositoryRoot(currentDirectory) {
			return currentDirectory, nil
		}

		parent := filepath.Dir(currentDirectory)
		if parent == currentDirectory {
			return "", errors.New(
				"could not locate the GroundedSearch repository root",
			)
		}
		currentDirectory = parent
	}
}

func isRepositoryRoot(directory string) bool {
	required := []string{
		filepath.Join(directory, "apps", "search-api", "go.mod"),
		filepath.Join(directory, "services", "ml-service", "pyproject.toml"),
		filepath.Join(directory, "api", "internal", "ml-service-v1.yaml"),
	}

	for _, path := range required {
		information, err := os.Stat(path)
		if err != nil || information.IsDir() {
			return false
		}
	}

	return true
}

func selectPython(repositoryRoot string) (string, error) {
	if configured := strings.TrimSpace(os.Getenv("GROUNDEDSEARCH_PYTHON")); configured != "" {
		information, err := os.Stat(configured)
		if err != nil {
			return "", fmt.Errorf(
				"configured GROUNDEDSEARCH_PYTHON %q is unavailable: %w",
				configured,
				err,
			)
		}
		if information.IsDir() {
			return "", fmt.Errorf(
				"configured GROUNDEDSEARCH_PYTHON %q is a directory",
				configured,
			)
		}
		return configured, nil
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
		information, err := os.Stat(candidate)
		if err == nil && !information.IsDir() {
			return candidate, nil
		}
	}

	python, err := exec.LookPath("python")
	if err == nil {
		return python, nil
	}

	python3, err := exec.LookPath("python3")
	if err == nil {
		return python3, nil
	}

	return "", errors.New(
		"Python was not found in the ML service virtual environment or PATH",
	)
}

func pythonEnvironment(repositoryRoot string) []string {
	return setEnvironment(
		os.Environ(),
		"PYTHONPATH",
		filepath.Join(repositoryRoot, "services", "ml-service", "src"),
	)
}

func integrationEnvironment(repositoryRoot string, python string) []string {
	environment := setEnvironment(os.Environ(), "GROUNDEDSEARCH_PYTHON", python)
	return setEnvironment(
		environment,
		"PYTHONPATH",
		filepath.Join(repositoryRoot, "services", "ml-service", "src"),
	)
}

func setEnvironment(environment []string, key string, value string) []string {
	prefix := strings.ToUpper(key) + "="
	result := make([]string, 0, len(environment)+1)

	for _, entry := range environment {
		if strings.HasPrefix(strings.ToUpper(entry), prefix) {
			continue
		}
		result = append(result, entry)
	}

	return append(result, key+"="+value)
}

func formatCommand(name string, args []string) string {
	parts := append([]string{name}, args...)
	for index, part := range parts {
		if strings.ContainsAny(part, " \t\n\"") {
			parts[index] = fmt.Sprintf("%q", part)
		}
	}
	return strings.Join(parts, " ")
}
