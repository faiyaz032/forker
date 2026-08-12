package runtime

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseRunArgsDefaults(t *testing.T) {
	cfg, err := parseRunArgs([]string{"/bin/sh"})
	if err != nil {
		t.Fatalf("parseRunArgs() error = %v", err)
	}

	if cfg.MemoryMax != "256M" {
		t.Errorf("MemoryMax = %q, want %q", cfg.MemoryMax, "256M")
	}
	if cfg.CPUQuota != 1.0 {
		t.Errorf("CPUQuota = %v, want %v", cfg.CPUQuota, 1.0)
	}
	if cfg.PidsMax != 256 {
		t.Errorf("PidsMax = %d, want %d", cfg.PidsMax, 256)
	}
	if !reflect.DeepEqual(cfg.Command, []string{"/bin/sh"}) {
		t.Errorf("Command = %v, want %v", cfg.Command, []string{"/bin/sh"})
	}
}

func TestParseRunArgsEmpty(t *testing.T) {
	cfg, err := parseRunArgs(nil)
	if err != nil {
		t.Fatalf("parseRunArgs() error = %v", err)
	}

	if len(cfg.Command) != 0 {
		t.Errorf("Command = %v, want empty", cfg.Command)
	}
	if cfg.MemoryMax != "256M" || cfg.CPUQuota != 1.0 || cfg.PidsMax != 256 {
		t.Errorf("cfg = %+v, want defaults preserved", cfg)
	}
}

func TestParseRunArgsFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		memory  string
		cpu     float64
		pids    int
		command []string
	}{
		{
			name:    "memory only",
			args:    []string{"--memory", "512M", "/bin/sh"},
			memory:  "512M",
			cpu:     1.0,
			pids:    256,
			command: []string{"/bin/sh"},
		},
		{
			name:    "cpu only",
			args:    []string{"--cpu", "0.5", "/bin/sh"},
			memory:  "256M",
			cpu:     0.5,
			pids:    256,
			command: []string{"/bin/sh"},
		},
		{
			name:    "pids only",
			args:    []string{"--pids", "64", "/bin/sh"},
			memory:  "256M",
			cpu:     1.0,
			pids:    64,
			command: []string{"/bin/sh"},
		},
		{
			name:    "all flags",
			args:    []string{"--memory", "1G", "--cpu", "2", "--pids", "10", "/bin/sh", "-c", "echo hi"},
			memory:  "1G",
			cpu:     2,
			pids:    10,
			command: []string{"/bin/sh", "-c", "echo hi"},
		},
		{
			name:    "double dash separator",
			args:    []string{"--memory", "64M", "--", "/bin/ls", "-la"},
			memory:  "64M",
			cpu:     1.0,
			pids:    256,
			command: []string{"/bin/ls", "-la"},
		},
		{
			name:    "double dash with nothing after",
			args:    []string{"--"},
			memory:  "256M",
			cpu:     1.0,
			pids:    256,
			command: []string{},
		},
		{
			name:    "command flags are not consumed",
			args:    []string{"--", "/bin/ls", "--memory", "512M"},
			memory:  "256M",
			cpu:     1.0,
			pids:    256,
			command: []string{"/bin/ls", "--memory", "512M"},
		},
		{
			name:    "later flag wins",
			args:    []string{"--cpu", "1", "--cpu", "4", "/bin/sh"},
			memory:  "256M",
			cpu:     4,
			pids:    256,
			command: []string{"/bin/sh"},
		},
		{
			name:    "zero cpu",
			args:    []string{"--cpu", "0", "/bin/sh"},
			memory:  "256M",
			cpu:     0,
			pids:    256,
			command: []string{"/bin/sh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseRunArgs(tt.args)
			if err != nil {
				t.Fatalf("parseRunArgs() error = %v", err)
			}

			if cfg.MemoryMax != tt.memory {
				t.Errorf("MemoryMax = %q, want %q", cfg.MemoryMax, tt.memory)
			}
			if cfg.CPUQuota != tt.cpu {
				t.Errorf("CPUQuota = %v, want %v", cfg.CPUQuota, tt.cpu)
			}
			if cfg.PidsMax != tt.pids {
				t.Errorf("PidsMax = %d, want %d", cfg.PidsMax, tt.pids)
			}
			if !reflect.DeepEqual(cfg.Command, tt.command) {
				t.Errorf("Command = %v, want %v", cfg.Command, tt.command)
			}
		})
	}
}

func TestParseRunArgsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"non numeric cpu", []string{"--cpu", "fast", "/bin/sh"}},
		{"empty cpu", []string{"--cpu", "", "/bin/sh"}},
		{"non numeric pids", []string{"--pids", "many", "/bin/sh"}},
		{"fractional pids", []string{"--pids", "1.5", "/bin/sh"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseRunArgs(tt.args); err == nil {
				t.Errorf("parseRunArgs(%v) = nil error, want error", tt.args)
			}
		})
	}
}

func TestParseRunArgsMissingFlagValuePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("parseRunArgs() did not panic on a trailing flag with no value")
		}
	}()

	_, _ = parseRunArgs([]string{"--memory"})
}

func TestIsChildProcess(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"1", true},
		{"0", false},
		{"", false},
		{"true", false},
	}

	for _, tt := range tests {
		t.Run("value_"+tt.value, func(t *testing.T) {
			t.Setenv(childEnv, tt.value)

			if got := IsChildProcess(); got != tt.want {
				t.Errorf("IsChildProcess() with %q = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsChildProcessUnset(t *testing.T) {
	t.Setenv(childEnv, "1")
	os.Unsetenv(childEnv)

	if IsChildProcess() {
		t.Error("IsChildProcess() = true, want false when env is unset")
	}
}

func TestUsage(t *testing.T) {
	stdout, _ := captureOutput(t, Usage)

	if !strings.Contains(stdout, "forker run") {
		t.Errorf("stdout = %q, want usage line", stdout)
	}
}

func TestRunArgErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"no args at all", []string{"forker"}, "invalid args"},
		{"empty", nil, "invalid args"},
		{"unknown command", []string{"forker", "bogus"}, "unknown command"},
		{"run without command", []string{"forker", "run"}, "missing command"},
		{"run with only flags", []string{"forker", "run", "--"}, "missing command"},
		{"stop without id", []string{"forker", "stop"}, "missing sandbox id"},
		{"exec without id", []string{"forker", "exec"}, "usage: forker exec"},
		{"exec without command", []string{"forker", "exec", "forker-1a2b"}, "usage: forker exec"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			captureOutput(t, func() {
				err = Run(tt.args)
			})

			if err == nil {
				t.Fatalf("Run(%v) = nil, want error", tt.args)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Run(%v) error = %q, want it to contain %q", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestRunPropagatesParseError(t *testing.T) {
	var err error
	captureOutput(t, func() {
		err = Run([]string{"forker", "run", "--cpu", "abc", "/bin/sh"})
	})

	if err == nil {
		t.Fatal("Run() = nil, want parse error")
	}
	if strings.Contains(err.Error(), "missing command") {
		t.Errorf("Run() error = %q, want the parse failure, not a missing command", err)
	}
}

func TestRunPsUsesBasePath(t *testing.T) {
	tempBasePath(t)

	var err error
	stdout, _ := captureOutput(t, func() {
		err = Run([]string{"forker", "ps"})
	})

	if err != nil {
		t.Fatalf("Run(ps) = %v, want nil", err)
	}
	if !strings.Contains(stdout, "STATUS") {
		t.Errorf("stdout = %q, want a table header", stdout)
	}
}

func TestRunStopUnknownSandbox(t *testing.T) {
	tempBasePath(t)

	var err error
	captureOutput(t, func() {
		err = Run([]string{"forker", "stop", "forker-nope"})
	})

	if err == nil {
		t.Fatal("Run(stop) = nil, want error for unknown sandbox")
	}
	if !strings.Contains(err.Error(), "cannot read sandbox") {
		t.Errorf("Run(stop) error = %q, want read failure", err)
	}
}

func TestRunUnknownCommandPrintsUsage(t *testing.T) {
	stdout, _ := captureOutput(t, func() {
		_ = Run([]string{"forker", "bogus"})
	})

	if !strings.Contains(stdout, "forker run") {
		t.Errorf("stdout = %q, want usage printed", stdout)
	}
}

func TestRunInNamespaceUnknownBinary(t *testing.T) {
	cfg := RunConfig{Command: []string{"forker-definitely-not-a-real-binary"}}

	err := runInNamespace(cfg)
	if err == nil {
		t.Fatal("runInNamespace() = nil, want lookup error")
	}
	if !strings.Contains(err.Error(), "cannot find") {
		t.Errorf("err = %q, want lookup failure", err)
	}
}

func TestRunInNamespaceCleansUpAfterFailedStart(t *testing.T) {
	skipIfRoot(t)

	dir := tempBasePath(t)

	var err error
	stdout, _ := captureOutput(t, func() {
		err = runInNamespace(RunConfig{Command: []string{"true"}})
	})

	if err == nil {
		t.Fatal("runInNamespace() = nil, want error when namespaces are unavailable")
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("readdir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Errorf("base path holds %d entries after a failed start, want the sandbox state torn down", len(entries))
	}
	if !strings.Contains(stdout, "cleaning up veth") {
		t.Errorf("stdout = %q, want the veth teardown to run", stdout)
	}
}

func TestRunInNamespaceResolvesCommandFromPath(t *testing.T) {
	skipIfRoot(t)

	tempBasePath(t)

	var err error
	stdout, _ := captureOutput(t, func() {
		err = runInNamespace(RunConfig{Command: []string{"true"}})
	})

	if err == nil {
		t.Fatal("runInNamespace() = nil, want error when namespaces are unavailable")
	}
	if !strings.Contains(stdout, "starting") || !strings.Contains(stdout, "/true") {
		t.Errorf("stdout = %q, want the resolved absolute binary path", stdout)
	}
	if !strings.Contains(stdout, "in sandbox forker-") {
		t.Errorf("stdout = %q, want a generated sandbox id", stdout)
	}
}

func TestSaveSandbox(t *testing.T) {
	dir := tempBasePath(t)
	id := "forker-1a2b"

	if err := saveSandbox(id, 4242, "/bin/sh"); err != nil {
		t.Fatalf("saveSandbox() = %v, want nil", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, id, "pid"))
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	if string(data) != "4242" {
		t.Errorf("pid file = %q, want %q", data, "4242")
	}
}

func TestSaveSandboxOverwrites(t *testing.T) {
	dir := tempBasePath(t)
	id := "forker-1a2b"

	if err := saveSandbox(id, 1, "/bin/sh"); err != nil {
		t.Fatalf("saveSandbox() = %v", err)
	}
	if err := saveSandbox(id, 999, "/bin/sh"); err != nil {
		t.Fatalf("saveSandbox() = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, id, "pid"))
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	if string(data) != "999" {
		t.Errorf("pid file = %q, want %q", data, "999")
	}
}

func TestSaveSandboxFailsWhenBasePathIsAFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "not-a-dir")

	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	withBasePath(t, file)

	if err := saveSandbox("forker-1a2b", 1, "/bin/sh"); err == nil {
		t.Fatal("saveSandbox() = nil, want error when basePath is a file")
	}
}

func TestRunConfigZeroValue(t *testing.T) {
	var cfg RunConfig

	if cfg.Command != nil || cfg.MemoryMax != "" || cfg.CPUQuota != 0 || cfg.PidsMax != 0 {
		t.Errorf("zero RunConfig = %+v, want all fields empty", cfg)
	}
}
