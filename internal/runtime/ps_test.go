package runtime

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func writeSandbox(t *testing.T, dir, id, pid string) {
	t.Helper()

	sandboxDir := filepath.Join(dir, id)
	if err := os.MkdirAll(sandboxDir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", sandboxDir, err)
	}
	if pid == "" {
		return
	}
	if err := os.WriteFile(filepath.Join(sandboxDir, "pid"), []byte(pid), 0644); err != nil {
		t.Fatalf("write pid: %v", err)
	}
}

func TestIsAliveSelf(t *testing.T) {
	if !isAlive(os.Getpid()) {
		t.Error("isAlive(self) = false, want true")
	}
}

func TestIsAliveInit(t *testing.T) {
	if !isAlive(1) {
		t.Error("isAlive(1) = false, want true")
	}
}

func TestIsAliveDeadPids(t *testing.T) {
	tests := []struct {
		name string
		pid  int
	}{
		{"zero", 0},
		{"negative", -1},
		{"beyond pid max", 1 << 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isAlive(tt.pid) {
				t.Errorf("isAlive(%d) = true, want false", tt.pid)
			}
		})
	}
}

func TestListSandboxesEmpty(t *testing.T) {
	tempBasePath(t)

	var err error
	stdout, _ := captureOutput(t, func() {
		err = listSandboxes()
	})

	if err != nil {
		t.Fatalf("listSandboxes() = %v, want nil", err)
	}
	if !strings.Contains(stdout, "ID") || !strings.Contains(stdout, "STATUS") {
		t.Errorf("stdout = %q, want header", stdout)
	}
	if strings.Count(strings.TrimSpace(stdout), "\n") != 0 {
		t.Errorf("stdout = %q, want header only", stdout)
	}
}

func TestListSandboxesMissingBasePath(t *testing.T) {
	withBasePath(t, filepath.Join(t.TempDir(), "does-not-exist"))

	var err error
	captureOutput(t, func() {
		err = listSandboxes()
	})

	if err == nil {
		t.Fatal("listSandboxes() = nil, want error for missing base path")
	}
}

func TestListSandboxesRunning(t *testing.T) {
	dir := tempBasePath(t)
	writeSandbox(t, dir, "forker-live", strconv.Itoa(os.Getpid()))

	var err error
	stdout, _ := captureOutput(t, func() {
		err = listSandboxes()
	})

	if err != nil {
		t.Fatalf("listSandboxes() = %v, want nil", err)
	}
	if !strings.Contains(stdout, "forker-live") {
		t.Errorf("stdout = %q, want sandbox id", stdout)
	}
	if !strings.Contains(stdout, "Running") {
		t.Errorf("stdout = %q, want Running status", stdout)
	}
}

func TestListSandboxesStopped(t *testing.T) {
	dir := tempBasePath(t)
	writeSandbox(t, dir, "forker-dead", strconv.Itoa(1<<30))

	stdout, _ := captureOutput(t, func() {
		_ = listSandboxes()
	})

	if !strings.Contains(stdout, "Stopped") {
		t.Errorf("stdout = %q, want Stopped status", stdout)
	}
}

func TestListSandboxesTrimsWhitespaceInPidFile(t *testing.T) {
	dir := tempBasePath(t)
	writeSandbox(t, dir, "forker-padded", "  "+strconv.Itoa(os.Getpid())+"\n")

	stdout, _ := captureOutput(t, func() {
		_ = listSandboxes()
	})

	if !strings.Contains(stdout, "Running") {
		t.Errorf("stdout = %q, want Running for a padded pid file", stdout)
	}
}

func TestListSandboxesMissingPidFile(t *testing.T) {
	dir := tempBasePath(t)
	writeSandbox(t, dir, "forker-nopid", "")

	var err error
	stdout, _ := captureOutput(t, func() {
		err = listSandboxes()
	})

	if err != nil {
		t.Fatalf("listSandboxes() = %v, want nil", err)
	}
	if !strings.Contains(stdout, "cannot read pid file for forker-nopid") {
		t.Errorf("stdout = %q, want a warning about the missing pid file", stdout)
	}
}

func TestListSandboxesInvalidPidFile(t *testing.T) {
	dir := tempBasePath(t)
	writeSandbox(t, dir, "forker-bad", "not-a-number")

	var err error
	stdout, _ := captureOutput(t, func() {
		err = listSandboxes()
	})

	if err != nil {
		t.Fatalf("listSandboxes() = %v, want nil", err)
	}
	if !strings.Contains(stdout, "invalid pid for forker-bad") {
		t.Errorf("stdout = %q, want a warning about the invalid pid", stdout)
	}
}

func TestListSandboxesSkipsFiles(t *testing.T) {
	dir := tempBasePath(t)

	if err := os.WriteFile(filepath.Join(dir, "stray.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	stdout, _ := captureOutput(t, func() {
		_ = listSandboxes()
	})

	if strings.Contains(stdout, "stray.txt") {
		t.Errorf("stdout = %q, want non-directory entries skipped", stdout)
	}
}

func TestListSandboxesContinuesAfterBadEntry(t *testing.T) {
	dir := tempBasePath(t)
	writeSandbox(t, dir, "forker-aaa", "not-a-number")
	writeSandbox(t, dir, "forker-bbb", strconv.Itoa(os.Getpid()))

	stdout, _ := captureOutput(t, func() {
		_ = listSandboxes()
	})

	if !strings.Contains(stdout, "invalid pid for forker-aaa") {
		t.Errorf("stdout = %q, want warning for the bad entry", stdout)
	}
	if !strings.Contains(stdout, "forker-bbb") {
		t.Errorf("stdout = %q, want the healthy entry still listed", stdout)
	}
}
