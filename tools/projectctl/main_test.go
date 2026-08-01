package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type recordedCommand struct {
	name        string
	args        []string
	directory   string
	environment []string
}

type recordingRunner struct {
	commands []recordedCommand
	failAt   int
}

func (runner *recordingRunner) Run(
	name string,
	args []string,
	directory string,
	environment []string,
) error {
	runner.commands = append(runner.commands, recordedCommand{
		name:        name,
		args:        append([]string(nil), args...),
		directory:   directory,
		environment: append([]string(nil), environment...),
	})

	if runner.failAt > 0 && len(runner.commands) == runner.failAt {
		return errors.New("test failure")
	}

	return nil
}

func TestExecuteHelp(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	app := application{stdout: &output}

	if err := app.execute("help"); err != nil {
		t.Fatalf("execute(help) error = %v", err)
	}
	if !strings.Contains(output.String(), "verify-core") {
		t.Fatalf("help output does not list verify-core:\n%s", output.String())
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	t.Parallel()

	app := application{stdout: &bytes.Buffer{}}
	if err := app.execute("unknown"); err == nil {
		t.Fatal("execute(unknown) error = nil, want error")
	}
}

func TestVerifyCoreRunsEveryStepInOrder(t *testing.T) {
	t.Parallel()

	repositoryRoot := t.TempDir()
	runner := &recordingRunner{}
	app := application{
		repositoryRoot: repositoryRoot,
		python:         "python-test",
		runner:         runner,
		stdout:         &bytes.Buffer{},
	}

	if err := app.verifyCore(); err != nil {
		t.Fatalf("verifyCore() error = %v", err)
	}

	if len(runner.commands) != 5 {
		t.Fatalf("command count = %d, want 5", len(runner.commands))
	}

	if runner.commands[0].name != "python-test" {
		t.Errorf("first command = %q, want python-test", runner.commands[0].name)
	}
	if runner.commands[1].name != "go" {
		t.Errorf("second command = %q, want go", runner.commands[1].name)
	}
	if runner.commands[4].directory != filepath.Join(repositoryRoot, "tests", "integration") {
		t.Errorf(
			"integration directory = %q",
			runner.commands[4].directory,
		)
	}
}

func TestVerifyCoreStopsAtFirstFailure(t *testing.T) {
	t.Parallel()

	runner := &recordingRunner{failAt: 2}
	app := application{
		repositoryRoot: t.TempDir(),
		python:         "python-test",
		runner:         runner,
		stdout:         &bytes.Buffer{},
	}

	if err := app.verifyCore(); err == nil {
		t.Fatal("verifyCore() error = nil, want error")
	}
	if len(runner.commands) != 2 {
		t.Fatalf("command count = %d, want 2", len(runner.commands))
	}
}

func TestSelectPythonUsesConfiguredExecutable(t *testing.T) {
	configured := filepath.Join(t.TempDir(), "configured-python")
	if runtime.GOOS == "windows" {
		configured += ".exe"
	}
	if err := os.WriteFile(configured, []byte("placeholder"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("GROUNDEDSEARCH_PYTHON", configured)

	selected, err := selectPython(t.TempDir())
	if err != nil {
		t.Fatalf("selectPython() error = %v", err)
	}
	if selected != configured {
		t.Fatalf("selected = %q, want %q", selected, configured)
	}
}

func TestSelectPythonRejectsMissingConfiguredExecutable(t *testing.T) {
	configured := filepath.Join(t.TempDir(), "missing-python")
	t.Setenv("GROUNDEDSEARCH_PYTHON", configured)

	if _, err := selectPython(t.TempDir()); err == nil {
		t.Fatal("selectPython() error = nil, want error")
	}
}

func TestSelectPythonPrefersVirtualEnvironment(t *testing.T) {
	t.Setenv("GROUNDEDSEARCH_PYTHON", "")

	if runtime.GOOS == "windows" {
		t.Skip("POSIX virtual-environment path test")
	}

	repositoryRoot := t.TempDir()
	pythonPath := filepath.Join(
		repositoryRoot,
		"services",
		"ml-service",
		".venv",
		"bin",
		"python",
	)
	if err := os.MkdirAll(filepath.Dir(pythonPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(pythonPath, []byte("placeholder"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	selected, err := selectPython(repositoryRoot)
	if err != nil {
		t.Fatalf("selectPython() error = %v", err)
	}
	if selected != pythonPath {
		t.Fatalf("selected = %q, want %q", selected, pythonPath)
	}
}

func TestSetEnvironmentReplacesExistingValueCaseInsensitively(t *testing.T) {
	t.Parallel()

	environment := []string{
		"PATH=/bin",
		"PythonPath=old",
	}
	updated := setEnvironment(environment, "PYTHONPATH", "new")

	matches := 0
	for _, entry := range updated {
		if strings.HasPrefix(strings.ToUpper(entry), "PYTHONPATH=") {
			matches++
			if entry != "PYTHONPATH=new" {
				t.Errorf("PYTHONPATH entry = %q, want PYTHONPATH=new", entry)
			}
		}
	}

	if matches != 1 {
		t.Fatalf("PYTHONPATH entry count = %d, want 1", matches)
	}
}

func TestFormatCommandQuotesWhitespace(t *testing.T) {
	t.Parallel()

	actual := formatCommand("python", []string{"path with spaces/script.py"})
	if actual != `python "path with spaces/script.py"` {
		t.Fatalf("formatCommand() = %q", actual)
	}
}
