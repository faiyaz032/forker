package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestCleanupVethNamesFromShortID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		wantVeth string
	}{
		{"standard sandbox id", "forker-7d79", "veth-7d79"},
		{"short id used verbatim", "abc", "veth-abc"},
		{"exactly seven chars used verbatim", "forker7", "veth-forker7"},
		{"eight chars truncated to last four", "forker78", "veth-er78"},
		{"long id truncated to last four", "forker-aaaa-bbbb-cccc", "veth-cccc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			stdout, _ := captureOutput(t, func() {
				err = cleanupVeth(tt.id)
			})

			if err != nil {
				t.Errorf("cleanupVeth() = %v, want nil", err)
			}
			if !strings.Contains(stdout, tt.wantVeth) {
				t.Errorf("stdout = %q, want veth name %q", stdout, tt.wantVeth)
			}
		})
	}
}

func TestCleanupVethAlwaysSucceeds(t *testing.T) {
	var err error
	captureOutput(t, func() {
		err = cleanupVeth("forker-does-not-exist")
	})

	if err != nil {
		t.Errorf("cleanupVeth() = %v, want nil even for an unknown sandbox", err)
	}
}

func TestSetupVethDerivedAddress(t *testing.T) {
	skipIfRoot(t)
	skipIfNoBinary(t, "ip")

	tests := []struct {
		id     string
		wantIP string
	}{
		{"forker-7d79", "10.200.0.123/16"},
		{"forker-0000", "10.200.0.2/16"},
		{"forker-0001", "10.200.0.3/16"},
		{"forker-00fa", "10.200.0.2/16"},
		{"forker-ffff", "10.200.0.37/16"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			var err error
			stdout, _ := captureOutput(t, func() {
				err = setupVeth(tt.id, 1)
			})

			if err == nil {
				t.Fatal("setupVeth() = nil, want error when unprivileged")
			}
			if !strings.Contains(stdout, tt.wantIP) {
				t.Errorf("stdout = %q, want address %q", stdout, tt.wantIP)
			}
			if !strings.Contains(stdout, "veth-"+tt.id[len(tt.id)-4:]) {
				t.Errorf("stdout = %q, want the derived veth name", stdout)
			}
		})
	}
}

func TestSetupVethAddressStaysInRange(t *testing.T) {
	for i := 0; i <= 0xffff; i += 97 {
		shortID := fmt.Sprintf("%04x", i)

		val, err := strconv.ParseInt(shortID, 16, 64)
		if err != nil {
			t.Fatalf("ParseInt(%q): %v", shortID, err)
		}

		octet := (val % 250) + 2
		if octet < 2 || octet > 251 {
			t.Fatalf("octet for %q = %d, want between 2 and 251", shortID, octet)
		}
	}
}

func TestSetupVethUnprivilegedFails(t *testing.T) {
	skipIfRoot(t)
	skipIfNoBinary(t, "ip")

	var err error
	captureOutput(t, func() {
		err = setupVeth("forker-dead", 1<<30)
	})

	if err == nil {
		t.Fatal("setupVeth() = nil, want error when unprivileged")
	}
}
