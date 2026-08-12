package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCPUToCgroupQuota(t *testing.T) {
	tests := []struct {
		name string
		cpu  float64
		want string
	}{
		{"zero is unlimited", 0, "max"},
		{"negative is unlimited", -1.5, "max"},
		{"one full core", 1.0, "100000 100000"},
		{"half a core", 0.5, "50000 100000"},
		{"two and a half cores", 2.5, "250000 100000"},
		{"four cores", 4, "400000 100000"},
		{"tiny slice", 0.001, "100 100000"},
		{"fractional rounds down", 0.30001, "30001 100000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cpuToCgroupQuota(tt.cpu); got != tt.want {
				t.Errorf("cpuToCgroupQuota(%v) = %q, want %q", tt.cpu, got, tt.want)
			}
		})
	}
}

func TestSetupCgroupsUnprivilegedFails(t *testing.T) {
	skipIfRoot(t)

	cfg := RunConfig{MemoryMax: "128M", CPUQuota: 1, PidsMax: 32}

	if err := setupCgroups("forker-test", os.Getpid(), cfg); err == nil {
		t.Fatal("setupCgroups() = nil, want error when unprivileged")
	}
}

func TestSetupCgroupsLeavesNoHostState(t *testing.T) {
	skipIfRoot(t)

	id := "forker-test-nostate"
	_ = setupCgroups(id, os.Getpid(), RunConfig{})

	if _, err := os.Stat(filepath.Join("/sys/fs/cgroup", "forker", id)); err == nil {
		t.Errorf("cgroup dir for %q exists, want none created when unprivileged", id)
	}
}
