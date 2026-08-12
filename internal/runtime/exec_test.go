package runtime

import (
	"os"
	"strings"
	"testing"
)

func TestExecChildInvalidArgsJSON(t *testing.T) {
	t.Setenv("__FORKER_CMD__", "true")
	t.Setenv("__FORKER_ARGS__", "not-json")

	if err := ExecChild(); err == nil {
		t.Fatal("ExecChild() = nil, want JSON error")
	}
}

func TestExecChildEmptyArgsEnv(t *testing.T) {
	t.Setenv("__FORKER_CMD__", "forker-definitely-not-a-real-binary")
	os.Unsetenv("__FORKER_ARGS__")

	if err := ExecChild(); err == nil {
		t.Fatal("ExecChild() = nil, want JSON error for an unset args env")
	}
}

func TestExecChildUnknownCommand(t *testing.T) {
	t.Setenv("__FORKER_CMD__", "forker-definitely-not-a-real-binary")
	t.Setenv("__FORKER_ARGS__", `[]`)

	err := ExecChild()
	if err == nil {
		t.Fatal("ExecChild() = nil, want lookup error")
	}
	if !strings.Contains(err.Error(), "forker-definitely-not-a-real-binary") {
		t.Errorf("err = %q, want it to name the command", err)
	}
}

func TestExecChildEmptyCommand(t *testing.T) {
	t.Setenv("__FORKER_CMD__", "")
	t.Setenv("__FORKER_ARGS__", `["-c","echo hi"]`)

	if err := ExecChild(); err == nil {
		t.Fatal("ExecChild() = nil, want lookup error for an empty command")
	}
}

func TestExecInSandboxUnknownSandbox(t *testing.T) {
	err := execInSandboxQuietly("forker-no-such-sandbox", "true", []string{})
	if err == nil {
		t.Fatal("execInSandboxQuietly() = nil, want failure entering a missing sandbox")
	}
}

func TestExecInSandboxQuietlySuppressesOutput(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		_ = execInSandboxQuietly("forker-no-such-sandbox", "true", []string{})
	})

	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestExecInSandboxVerboseForwardsStderr(t *testing.T) {
	var err error
	_, stderr := captureOutput(t, func() {
		err = execInSandbox("forker-no-such-sandbox", "true", []string{})
	})

	if err == nil {
		t.Fatal("execInSandbox() = nil, want failure entering a missing sandbox")
	}
	if !strings.Contains(stderr, "nsenter") {
		t.Errorf("stderr = %q, want the namespace entry failure forwarded", stderr)
	}
}

func TestExecInSandboxUnmarshalableArgs(t *testing.T) {
	err := execInSandboxQuietly("forker-no-such-sandbox", "true", nil)
	if err == nil {
		t.Fatal("execInSandboxQuietly() = nil, want failure entering a missing sandbox")
	}
}
