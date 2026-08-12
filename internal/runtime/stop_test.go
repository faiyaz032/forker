package runtime

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestStopSandboxMissingPidFile(t *testing.T) {
	tempBasePath(t)

	var err error
	captureOutput(t, func() {
		err = stopSandbox("forker-missing")
	})

	if err == nil {
		t.Fatal("stopSandbox() = nil, want error")
	}
	if !strings.Contains(err.Error(), "cannot read sandbox") {
		t.Errorf("err = %q, want read failure", err)
	}
}

func TestStopSandboxInvalidPid(t *testing.T) {
	dir := tempBasePath(t)
	writeSandbox(t, dir, "forker-bad", "garbage")

	var err error
	captureOutput(t, func() {
		err = stopSandbox("forker-bad")
	})

	if err == nil {
		t.Fatal("stopSandbox() = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid pid") {
		t.Errorf("err = %q, want invalid pid failure", err)
	}
}

func TestStopSandboxEmptyPidFile(t *testing.T) {
	dir := tempBasePath(t)
	writeSandbox(t, dir, "forker-empty", "\n")

	var err error
	captureOutput(t, func() {
		err = stopSandbox("forker-empty")
	})

	if err == nil {
		t.Fatal("stopSandbox() = nil, want error for an empty pid file")
	}
}

func TestStopSandboxAlreadyDead(t *testing.T) {
	dir := tempBasePath(t)
	id := "forker-gone"
	writeSandbox(t, dir, id, strconv.Itoa(1<<30))

	var err error
	stdout, _ := captureOutput(t, func() {
		err = stopSandbox(id)
	})

	if err != nil {
		t.Fatalf("stopSandbox() = %v, want nil", err)
	}
	if !strings.Contains(stdout, "stopped") {
		t.Errorf("stdout = %q, want confirmation", stdout)
	}
	if _, statErr := os.Stat(filepath.Join(dir, id)); !os.IsNotExist(statErr) {
		t.Errorf("sandbox dir still present: %v", statErr)
	}
}

func TestStopSandboxCleansUpStateFiles(t *testing.T) {
	dir := tempBasePath(t)
	id := "forker-clean"
	writeSandbox(t, dir, id, strconv.Itoa(1<<30))

	if err := os.WriteFile(filepath.Join(dir, id, "ready"), []byte("1"), 0644); err != nil {
		t.Fatalf("write ready: %v", err)
	}

	captureOutput(t, func() {
		_ = stopSandbox(id)
	})

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("base path still holds %d entries, want 0", len(entries))
	}
}

func TestStopSandboxKillsLiveProcessGroup(t *testing.T) {
	dir := tempBasePath(t)
	id := "forker-live-kill"

	cmd := exec.Command("sh", "-c", "sleep 30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	pid := cmd.Process.Pid
	waited := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(waited)
	}()

	writeSandbox(t, dir, id, strconv.Itoa(pid))

	var err error
	captureOutput(t, func() {
		err = stopSandbox(id)
	})

	if err != nil {
		t.Fatalf("stopSandbox() = %v, want nil", err)
	}

	select {
	case <-waited:
	case <-time.After(5 * time.Second):
		_ = syscall.Kill(-pid, syscall.SIGKILL)
		t.Fatal("process still running after stopSandbox()")
	}

	if _, statErr := os.Stat(filepath.Join(dir, id)); !os.IsNotExist(statErr) {
		t.Errorf("sandbox dir still present: %v", statErr)
	}
}

func TestStopSandboxUnsignalableGroupReturnsError(t *testing.T) {
	dir := tempBasePath(t)
	id := "forker-nogroup"

	cmd := exec.Command("sh", "-c", "sleep 30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	pid := cmd.Process.Pid
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	writeSandbox(t, dir, id, strconv.Itoa(pid))

	var err error
	captureOutput(t, func() {
		err = stopSandbox(id)
	})

	if err == nil {
		t.Fatal("stopSandbox() = nil, want error when the signal is refused")
	}
	if !strings.Contains(err.Error(), "cannot kill sandbox") {
		t.Errorf("err = %q, want kill failure", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, id)); statErr != nil {
		t.Errorf("sandbox state removed despite the failure: %v", statErr)
	}
}
