package runtime

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestNetworkConstants(t *testing.T) {
	if bridgeName != "forker0" {
		t.Errorf("bridgeName = %q, want %q", bridgeName, "forker0")
	}
	if subnetCIDR != "10.200.0.1/16" {
		t.Errorf("subnetCIDR = %q, want %q", subnetCIDR, "10.200.0.1/16")
	}
	if subnet != "10.200.0.0/16" {
		t.Errorf("subnet = %q, want %q", subnet, "10.200.0.0/16")
	}
	if !strings.HasPrefix(subnetCIDR, "10.200.") {
		t.Errorf("subnetCIDR %q is not inside subnet %q", subnetCIDR, subnet)
	}
}

func TestRunSuccess(t *testing.T) {
	var err error
	captureOutput(t, func() {
		err = run("sh", "-c", "exit 0")
	})

	if err != nil {
		t.Errorf("run() = %v, want nil", err)
	}
}

func TestRunForwardsOutput(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		_ = run("sh", "-c", "echo out; echo err 1>&2")
	})

	if !strings.Contains(stdout, "out") {
		t.Errorf("stdout = %q, want forwarded stdout", stdout)
	}
	if !strings.Contains(stderr, "err") {
		t.Errorf("stderr = %q, want forwarded stderr", stderr)
	}
}

func TestRunNonZeroExit(t *testing.T) {
	var err error
	captureOutput(t, func() {
		err = run("sh", "-c", "exit 3")
	})

	if err == nil {
		t.Fatal("run() = nil, want exit error")
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("run() error = %T, want *exec.ExitError", err)
	}
	if exitErr.ExitCode() != 3 {
		t.Errorf("exit code = %d, want 3", exitErr.ExitCode())
	}
}

func TestRunMissingBinary(t *testing.T) {
	var err error
	captureOutput(t, func() {
		err = run("forker-definitely-not-a-real-binary")
	})

	if err == nil {
		t.Fatal("run() = nil, want lookup error")
	}
}

func TestRunQuietlySuppressesOutput(t *testing.T) {
	var err error
	stdout, stderr := captureOutput(t, func() {
		err = runQuietly("sh", "-c", "echo out; echo err 1>&2")
	})

	if err != nil {
		t.Errorf("runQuietly() = %v, want nil", err)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRunQuietlyNonZeroExit(t *testing.T) {
	if err := runQuietly("sh", "-c", "exit 1"); err == nil {
		t.Error("runQuietly() = nil, want exit error")
	}
}

func TestLinkExistsLoopback(t *testing.T) {
	skipIfNoBinary(t, "ip")

	if !linkExists("lo") {
		t.Error("linkExists(lo) = false, want true")
	}
}

func TestLinkExistsUnknownInterface(t *testing.T) {
	skipIfNoBinary(t, "ip")

	if linkExists("forker-no-such-link") {
		t.Error("linkExists() = true, want false for an unknown interface")
	}
}

func TestSetupNetworkingUnprivilegedFails(t *testing.T) {
	skipIfRoot(t)
	skipIfNoBinary(t, "ip")

	var err error
	captureOutput(t, func() {
		err = setupNetworking(Config{SandboxID: "forker-test"})
	})

	if err == nil {
		t.Fatal("setupNetworking() = nil, want error when unprivileged")
	}
}

func TestInitNetworkUnprivilegedFails(t *testing.T) {
	skipIfRoot(t)
	skipIfNoBinary(t, "ip")

	var err error
	captureOutput(t, func() {
		err = initNetwork()
	})

	if err == nil {
		t.Fatal("initNetwork() = nil, want error when unprivileged")
	}
}

func TestInitNetworkCreatesNoBridgeWhenUnprivileged(t *testing.T) {
	skipIfRoot(t)
	skipIfNoBinary(t, "ip")

	if linkExists(bridgeName) {
		t.Skipf("skipping: bridge %q already exists on this host", bridgeName)
	}

	captureOutput(t, func() {
		_ = initNetwork()
	})

	if linkExists(bridgeName) {
		t.Errorf("bridge %q exists, want none created when unprivileged", bridgeName)
	}
}
