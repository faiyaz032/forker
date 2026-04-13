package runtime

import (
	"os"
	"os/exec"
)

func setupNetworking(cfg Config) error {
	ip, err := exec.LookPath("ip")
	if err != nil {
		return err
	}

	cmd := exec.Command(ip, "link", "set", "lo", "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
