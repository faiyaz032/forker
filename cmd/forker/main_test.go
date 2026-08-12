package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

func TestMain(m *testing.M) {
	code := m.Run()

	if binPath != "" {
		_ = os.RemoveAll(filepath.Dir(binPath))
	}
	os.Exit(code)
}

func forkerBinary(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "forker-build")
		if err != nil {
			buildErr = err
			return
		}

		binPath = filepath.Join(dir, "forker")

		cmd := exec.Command("go", "build", "-o", binPath, ".")
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = errors.New(string(out))
		}
	})

	if buildErr != nil {
		t.Fatalf("build forker: %v", buildErr)
	}
	return binPath
}

func runForker(t *testing.T, env []string, args ...string) (string, int) {
	t.Helper()

	cmd := exec.Command(forkerBinary(t), args...)
	cmd.Env = append(os.Environ(), env...)

	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("run forker: %v", err)
	}
	return string(out), exitErr.ExitCode()
}

func TestMainNoArgs(t *testing.T) {
	out, code := runForker(t, nil)

	if code != 1 {
		t.Errorf("exit code = %d, want 1; output: %s", code, out)
	}
	if !strings.Contains(out, "invalid args") {
		t.Errorf("output = %q, want the invalid args error", out)
	}
	if !strings.Contains(out, "forker run") {
		t.Errorf("output = %q, want usage", out)
	}
}

func TestMainUnknownCommand(t *testing.T) {
	out, code := runForker(t, nil, "bogus")

	if code != 1 {
		t.Errorf("exit code = %d, want 1; output: %s", code, out)
	}
	if !strings.Contains(out, "unknown command") {
		t.Errorf("output = %q, want the unknown command error", out)
	}
}

func TestMainErrorsArePrefixed(t *testing.T) {
	out, _ := runForker(t, nil, "stop")

	if !strings.Contains(out, "[forker] error:") {
		t.Errorf("output = %q, want the parent mode error prefix", out)
	}
	if !strings.Contains(out, "missing sandbox id") {
		t.Errorf("output = %q, want the underlying error", out)
	}
}

func TestMainRunWithoutCommand(t *testing.T) {
	out, code := runForker(t, nil, "run")

	if code != 1 {
		t.Errorf("exit code = %d, want 1; output: %s", code, out)
	}
	if !strings.Contains(out, "missing command") {
		t.Errorf("output = %q, want the missing command error", out)
	}
}

func TestMainExecWithoutCommand(t *testing.T) {
	out, code := runForker(t, nil, "exec", "forker-1a2b")

	if code != 1 {
		t.Errorf("exit code = %d, want 1; output: %s", code, out)
	}
	if !strings.Contains(out, "usage: forker exec") {
		t.Errorf("output = %q, want the exec usage error", out)
	}
}

func TestMainExecChildModeBadArgs(t *testing.T) {
	env := []string{"__FORKER_CMD__=true", "__FORKER_ARGS__=not-json"}
	out, code := runForker(t, env, "exec-child")

	if code != 1 {
		t.Errorf("exit code = %d, want 1; output: %s", code, out)
	}
	if !strings.Contains(out, "[forker exec-child] error:") {
		t.Errorf("output = %q, want the exec-child error prefix", out)
	}
}

func TestMainExecChildModeUnknownCommand(t *testing.T) {
	env := []string{"__FORKER_CMD__=forker-definitely-not-a-real-binary", "__FORKER_ARGS__=[]"}
	out, code := runForker(t, env, "exec-child")

	if code != 1 {
		t.Errorf("exit code = %d, want 1; output: %s", code, out)
	}
	if !strings.Contains(out, "[forker exec-child]") {
		t.Errorf("output = %q, want the exec-child error prefix", out)
	}
}

func TestMainExecChildModeTakesPrecedenceOverChildMode(t *testing.T) {
	env := []string{
		"__FORKER_CHILD__=1",
		"__FORKER_CMD__=forker-definitely-not-a-real-binary",
		"__FORKER_ARGS__=[]",
	}
	out, _ := runForker(t, env, "exec-child")

	if !strings.Contains(out, "[forker exec-child]") {
		t.Errorf("output = %q, want exec mode to win over child mode", out)
	}
	if strings.Contains(out, "child started") {
		t.Errorf("output = %q, want child mode not to run", out)
	}
}

func TestMainChildModeWithoutBinPanics(t *testing.T) {
	env := []string{"__FORKER_CHILD__=1"}
	out, code := runForker(t, env)

	if code != 2 {
		t.Errorf("exit code = %d, want 2 for a panic; output: %s", code, out)
	}
	if !strings.Contains(out, "__FORKER_BIN__ not set") {
		t.Errorf("output = %q, want the missing bin panic", out)
	}
}

func TestMainChildModeIgnoredWhenEnvIsNotOne(t *testing.T) {
	env := []string{"__FORKER_CHILD__=0"}
	out, code := runForker(t, env)

	if code != 1 {
		t.Errorf("exit code = %d, want 1; output: %s", code, out)
	}
	if !strings.Contains(out, "invalid args") {
		t.Errorf("output = %q, want parent mode dispatch", out)
	}
}

func TestMainPsSucceedsOrReportsMissingState(t *testing.T) {
	out, code := runForker(t, nil, "ps")

	if code == 0 {
		if !strings.Contains(out, "STATUS") {
			t.Errorf("output = %q, want a table header", out)
		}
		return
	}
	if !strings.Contains(out, "[forker] error:") {
		t.Errorf("output = %q, want an error prefix when state is unreadable", out)
	}
}
