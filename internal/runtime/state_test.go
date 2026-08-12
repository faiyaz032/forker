package runtime

import "testing"

func TestBasePathDefault(t *testing.T) {
	if basePath != "/var/run/forker" {
		t.Errorf("basePath = %q, want %q", basePath, "/var/run/forker")
	}
}

func TestSandboxFields(t *testing.T) {
	s := Sandbox{ID: "forker-1a2b", PID: 4242, Cmd: "/bin/sh"}

	if s.ID != "forker-1a2b" {
		t.Errorf("ID = %q, want %q", s.ID, "forker-1a2b")
	}
	if s.PID != 4242 {
		t.Errorf("PID = %d, want %d", s.PID, 4242)
	}
	if s.Cmd != "/bin/sh" {
		t.Errorf("Cmd = %q, want %q", s.Cmd, "/bin/sh")
	}
}

func TestSandboxZeroValue(t *testing.T) {
	var s Sandbox

	if s.ID != "" || s.PID != 0 || s.Cmd != "" {
		t.Errorf("zero Sandbox = %+v, want all fields empty", s)
	}
}
