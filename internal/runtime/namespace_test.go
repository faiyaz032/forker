package runtime

import (
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestSetHostnameUnprivilegedFails(t *testing.T) {
	skipIfRoot(t)

	var err error
	_, stderr := captureOutput(t, func() {
		err = setHostname(Config{Hostname: "forker-test"})
	})

	if err == nil {
		t.Fatal("setHostname() = nil, want error when unprivileged")
	}
	if !strings.Contains(stderr, "setHostname") {
		t.Errorf("stderr = %q, want it to mention the failing operation", stderr)
	}
}

func TestSetHostnameReturnsErrno(t *testing.T) {
	skipIfRoot(t)

	var err error
	captureOutput(t, func() {
		err = setHostname(Config{Hostname: "forker-test"})
	})

	if err != syscall.EPERM {
		t.Errorf("setHostname() = %v, want %v", err, syscall.EPERM)
	}
}

func TestSetHostnameDoesNotChangeHostHostname(t *testing.T) {
	skipIfRoot(t)

	before, err := os.Hostname()
	if err != nil {
		t.Fatalf("os.Hostname(): %v", err)
	}

	captureOutput(t, func() {
		_ = setHostname(Config{Hostname: "forker-should-not-apply"})
	})

	after, err := os.Hostname()
	if err != nil {
		t.Fatalf("os.Hostname(): %v", err)
	}
	if before != after {
		t.Errorf("hostname changed from %q to %q", before, after)
	}
}
