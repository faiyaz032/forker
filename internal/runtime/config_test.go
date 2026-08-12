package runtime

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"
)

const execModeGuardMessage = "[test] refusing to run the suite in exec mode"

func TestMain(m *testing.M) {
	if os.Getenv("__FORKER_EXEC__") == "1" {
		_, _ = os.Stderr.WriteString(execModeGuardMessage + "\n")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	origStdout := os.Stdout
	origStderr := os.Stderr

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	os.Stdout = outW
	os.Stderr = errW

	outCh := make(chan string, 1)
	errCh := make(chan string, 1)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, outR)
		outCh <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, errR)
		errCh <- buf.String()
	}()

	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	fn()

	_ = outW.Close()
	_ = errW.Close()

	return <-outCh, <-errCh
}

func withBasePath(t *testing.T, dir string) {
	t.Helper()

	orig := basePath
	basePath = dir
	t.Cleanup(func() {
		basePath = orig
	})
}

func tempBasePath(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	withBasePath(t, dir)
	return dir
}

func skipIfRoot(t *testing.T) {
	t.Helper()

	if os.Geteuid() == 0 {
		t.Skip("skipping: test asserts unprivileged failure but is running as root")
	}
}

func skipIfNoBinary(t *testing.T, name string) {
	t.Helper()

	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("skipping: %q not found in PATH", name)
	}
}

func TestConfigZeroValue(t *testing.T) {
	var cfg Config

	if cfg.Bin != "" {
		t.Errorf("Bin = %q, want empty", cfg.Bin)
	}
	if cfg.Args != nil {
		t.Errorf("Args = %v, want nil", cfg.Args)
	}
	if cfg.Hostname != "" {
		t.Errorf("Hostname = %q, want empty", cfg.Hostname)
	}
	if cfg.SandboxID != "" {
		t.Errorf("SandboxID = %q, want empty", cfg.SandboxID)
	}
}

func TestCaptureOutputHelper(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		_, _ = os.Stdout.WriteString("to-stdout")
		_, _ = os.Stderr.WriteString("to-stderr")
	})

	if stdout != "to-stdout" {
		t.Errorf("stdout = %q, want %q", stdout, "to-stdout")
	}
	if stderr != "to-stderr" {
		t.Errorf("stderr = %q, want %q", stderr, "to-stderr")
	}
}

func TestWithBasePathRestores(t *testing.T) {
	orig := basePath

	t.Run("override", func(t *testing.T) {
		dir := tempBasePath(t)
		if basePath != dir {
			t.Errorf("basePath = %q, want %q", basePath, dir)
		}
	})

	if basePath != orig {
		t.Errorf("basePath = %q after subtest, want restored %q", basePath, orig)
	}
}
