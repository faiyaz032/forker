package runtime

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runSelfWithEnv(t *testing.T, env ...string) (string, int) {
	t.Helper()

	self, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable(): %v", err)
	}

	cmd := exec.Command(self, "-test.run=^$")
	cmd.Env = append(os.Environ(), env...)

	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("run self: %v", err)
	}
	return string(out), exitErr.ExitCode()
}

func TestNsenterInertWithoutExecFlag(t *testing.T) {
	out, code := runSelfWithEnv(t)

	if code != 0 {
		t.Errorf("exit code = %d, want 0; output: %s", code, out)
	}
	if strings.Contains(out, "[nsenter]") {
		t.Errorf("output = %q, want no namespace entry attempt", out)
	}
}

func TestNsenterInertWhenExecFlagIsNotOne(t *testing.T) {
	out, code := runSelfWithEnv(t, "__FORKER_EXEC__=0", "__FORKER_ID__=forker-1a2b")

	if code != 0 {
		t.Errorf("exit code = %d, want 0; output: %s", code, out)
	}
	if strings.Contains(out, "[nsenter]") {
		t.Errorf("output = %q, want no namespace entry attempt", out)
	}
}

func TestNsenterRequiresSandboxID(t *testing.T) {
	out, code := runSelfWithEnv(t, "__FORKER_EXEC__=1")

	if code != 1 {
		t.Errorf("exit code = %d, want 1; output: %s", code, out)
	}
	if !strings.Contains(out, "__FORKER_ID__ not set") {
		t.Errorf("output = %q, want the missing id message", out)
	}
}

func TestNsenterMissingPidFile(t *testing.T) {
	out, code := runSelfWithEnv(t, "__FORKER_EXEC__=1", "__FORKER_ID__=forker-no-such-sandbox")

	if code != 1 {
		t.Errorf("exit code = %d, want 1; output: %s", code, out)
	}
	if !strings.Contains(out, "open pid file") {
		t.Errorf("output = %q, want the pid file failure", out)
	}
}

func TestNsenterRunsBeforeTestMain(t *testing.T) {
	out, _ := runSelfWithEnv(t, "__FORKER_EXEC__=1", "__FORKER_ID__=forker-no-such-sandbox")

	if strings.Contains(out, execModeGuardMessage) {
		t.Errorf("output = %q, want the constructor to exit before TestMain runs", out)
	}
}
