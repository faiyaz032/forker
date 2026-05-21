package runtime

import (
	"fmt"
	"os"
	"os/exec"
)

func setupNetworking(cfg Config) error {
	cmd := exec.Command("ip", "link", "set", "lo", "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

const (
	bridgeName = "forker0"
	subnetCIDR = "10.200.0.1/16"
	subnet     = "10.200.0.0/16"
)

func initNetwork() error {
	if !linkExists(bridgeName) {
		fmt.Println("[net] creating bridge:", bridgeName)
		if err := run("ip", "link", "add", bridgeName, "type", "bridge"); err != nil {
			return err
		}
	}

	if err := run("ip", "link", "set", bridgeName, "up"); err != nil {
		return err
	}

	// always try to add IP, ignore error if already exists
	_ = runQuietly("ip", "addr", "add", subnetCIDR, "dev", bridgeName)

	// enable ip forwarding
	if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644); err != nil {
		return err
	}

	// NAT - only add if not exists
	if err := runQuietly("iptables", "-t", "nat", "-C", "POSTROUTING", "-s", subnet, "!", "-o", bridgeName, "-j", "MASQUERADE"); err != nil {
		_ = runQuietly("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", subnet, "!", "-o", bridgeName, "-j", "MASQUERADE")
	}

	return nil
}

func linkExists(name string) bool {
	err := exec.Command("ip", "link", "show", name).Run()
	return err == nil
}

func run(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func runQuietly(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	return c.Run()
}
