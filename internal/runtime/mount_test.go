package runtime

import (
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestSetupMountsUnprivilegedFails(t *testing.T) {
	skipIfRoot(t)

	var err error
	_, stderr := captureOutput(t, func() {
		err = setupMounts()
	})

	if err == nil {
		t.Fatal("setupMounts() = nil, want error when unprivileged")
	}
	if !strings.Contains(stderr, "mount private /") {
		t.Errorf("stderr = %q, want it to name the first failing mount", stderr)
	}
}

func TestSetupMountsFailsOnPrivateRootBeforeTouchingProc(t *testing.T) {
	skipIfRoot(t)

	var err error
	captureOutput(t, func() {
		err = setupMounts()
	})

	if err != syscall.EPERM {
		t.Errorf("setupMounts() = %v, want raw %v from the private-root remount", err, syscall.EPERM)
	}

	if _, statErr := os.Stat("/proc/self/stat"); statErr != nil {
		t.Errorf("/proc looks unmounted after failed setupMounts: %v", statErr)
	}
}
