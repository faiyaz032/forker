package runtime

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("__FORKER_BIN__", "/bin/sh")
	t.Setenv("__FORKER_ARGS__", `["-c","echo hi"]`)
	t.Setenv("__FORKER_HOSTNAME__", "forker-1a2b")
	t.Setenv("__FORKER_SANDBOX_ID__", "forker-1a2b")

	cfg := loadConfig()

	want := Config{
		Bin:       "/bin/sh",
		Args:      []string{"-c", "echo hi"},
		Hostname:  "forker-1a2b",
		SandboxID: "forker-1a2b",
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("loadConfig() = %+v, want %+v", cfg, want)
	}
}

func TestLoadConfigEmptyArgs(t *testing.T) {
	t.Setenv("__FORKER_BIN__", "/bin/sh")
	t.Setenv("__FORKER_ARGS__", `[]`)
	t.Setenv("__FORKER_HOSTNAME__", "")
	t.Setenv("__FORKER_SANDBOX_ID__", "")

	cfg := loadConfig()

	if len(cfg.Args) != 0 {
		t.Errorf("Args = %v, want empty", cfg.Args)
	}
	if cfg.Hostname != "" || cfg.SandboxID != "" {
		t.Errorf("optional fields = %q/%q, want empty", cfg.Hostname, cfg.SandboxID)
	}
}

func TestLoadConfigIgnoresInvalidArgsJSON(t *testing.T) {
	t.Setenv("__FORKER_BIN__", "/bin/sh")
	t.Setenv("__FORKER_ARGS__", "not-json")

	cfg := loadConfig()

	if cfg.Args != nil {
		t.Errorf("Args = %v, want nil for unparseable JSON", cfg.Args)
	}
	if cfg.Bin != "/bin/sh" {
		t.Errorf("Bin = %q, want %q", cfg.Bin, "/bin/sh")
	}
}

func TestLoadConfigUnsetArgsEnv(t *testing.T) {
	t.Setenv("__FORKER_BIN__", "/bin/true")
	os.Unsetenv("__FORKER_ARGS__")

	cfg := loadConfig()

	if cfg.Args != nil {
		t.Errorf("Args = %v, want nil", cfg.Args)
	}
}

func TestLoadConfigPanicsWithoutBin(t *testing.T) {
	t.Setenv("__FORKER_BIN__", "")

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("loadConfig() did not panic with __FORKER_BIN__ unset")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "__FORKER_BIN__") {
			t.Errorf("panic = %v, want message naming __FORKER_BIN__", r)
		}
	}()

	loadConfig()
}

func TestMustEnv(t *testing.T) {
	t.Setenv("__FORKER_TEST_KEY__", "value")

	if got := mustEnv("__FORKER_TEST_KEY__"); got != "value" {
		t.Errorf("mustEnv() = %q, want %q", got, "value")
	}
}

func TestMustEnvPanicsOnEmpty(t *testing.T) {
	t.Setenv("__FORKER_TEST_EMPTY__", "")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("mustEnv() did not panic for empty value")
		}
	}()

	mustEnv("__FORKER_TEST_EMPTY__")
}

func TestMustEnvPanicsOnMissing(t *testing.T) {
	os.Unsetenv("__FORKER_TEST_MISSING__")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("mustEnv() did not panic for missing key")
		}
	}()

	mustEnv("__FORKER_TEST_MISSING__")
}

func TestWaitForSandboxReadyImmediate(t *testing.T) {
	dir := tempBasePath(t)
	id := "forker-ready"

	if err := os.MkdirAll(filepath.Join(dir, id), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, id, "ready"), []byte("1"), 0644); err != nil {
		t.Fatalf("write ready: %v", err)
	}

	if err := waitForSandboxReady(id); err != nil {
		t.Errorf("waitForSandboxReady() = %v, want nil", err)
	}
}

func TestWaitForSandboxReadyAfterDelay(t *testing.T) {
	dir := tempBasePath(t)
	id := "forker-delayed"

	if err := os.MkdirAll(filepath.Join(dir, id), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	go func() {
		time.Sleep(250 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(dir, id, "ready"), []byte("1"), 0644)
	}()

	if err := waitForSandboxReady(id); err != nil {
		t.Errorf("waitForSandboxReady() = %v, want nil", err)
	}
}

func TestWaitForSandboxReadyTimesOut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: waits out the full ~5s readiness window")
	}

	tempBasePath(t)

	err := waitForSandboxReady("forker-never")
	if err == nil {
		t.Fatal("waitForSandboxReady() = nil, want timeout error")
	}
	if !strings.Contains(err.Error(), "not ready") {
		t.Errorf("err = %v, want it to mention readiness", err)
	}
}

func TestChildMainPanicsWithoutBin(t *testing.T) {
	t.Setenv("__FORKER_BIN__", "")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("ChildMain() did not panic with __FORKER_BIN__ unset")
		}
	}()

	_ = ChildMain()
}

func TestChildMainUnprivilegedFailsAtHostname(t *testing.T) {
	skipIfRoot(t)

	t.Setenv("__FORKER_BIN__", "/bin/true")
	t.Setenv("__FORKER_ARGS__", `[]`)
	t.Setenv("__FORKER_HOSTNAME__", "forker-test")
	t.Setenv("__FORKER_SANDBOX_ID__", "forker-test")

	var err error
	stdout, _ := captureOutput(t, func() {
		err = ChildMain()
	})

	if err == nil {
		t.Fatal("ChildMain() = nil, want error when unprivileged")
	}
	if !strings.Contains(stdout, "child started") {
		t.Errorf("stdout = %q, want startup line", stdout)
	}
}
