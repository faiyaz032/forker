package runtime

import (
	"fmt"
	"strconv"
)

func setupVeth(id string, pid int) error {
	// Extract the unique hex part (e.g., "7d79" from "forker-7d79")
	shortID := id
	if len(id) > 7 {
		shortID = id[len(id)-4:]
	}

	vethHost := "veth-" + shortID
	vethNS := "veth-ns-" + shortID

	// get IP from the shortID to avoid resets on every CLI run
	// shortID is hex, e.g., "7d79"
	val, _ := strconv.ParseInt(shortID, 16, 64)
	octet := (val % 250) + 2 // 2-251
	ipAddr := fmt.Sprintf("10.200.0.%d/16", octet)

	fmt.Printf("[forker][net] wiring sandbox %s -> %s (veth: %s)\n", id, ipAddr, vethHost)

	// create veth pair
	_ = runQuietly("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethNS)

	// attach host side to bridge
	if err := run("ip", "link", "set", vethHost, "master", bridgeName); err != nil {
		return err
	}
	if err := run("ip", "link", "set", vethHost, "up"); err != nil {
		return err
	}

	// move ns side into container
	if err := run("ip", "link", "set", vethNS, "netns", strconv.Itoa(pid)); err != nil {
		return err
	}

	// configure inside sandbox via your exec system
	if err := execInSandbox(id, "ip", []string{"link", "set", "lo", "up"}); err != nil {
		return err
	}

	if err := execInSandbox(id, "ip", []string{"link", "set", vethNS, "name", "eth0"}); err != nil {
		return err
	}

	if err := execInSandbox(id, "ip", []string{"addr", "add", ipAddr, "dev", "eth0"}); err != nil {
		return err
	}

	if err := execInSandbox(id, "ip", []string{"link", "set", "eth0", "up"}); err != nil {
		return err
	}

	_ = execInSandboxQuietly(id, "ip", []string{"route", "add", "default", "via", "10.200.0.1"})

	return nil
}

func cleanupVeth(id string) error {

	shortID := id
	if len(id) > 7 {
		shortID = id[len(id)-4:]
	}
	vethHost := "veth-" + shortID

	fmt.Printf("[net] cleaning up veth: %s\n", vethHost)
	_ = runQuietly("ip", "link", "delete", vethHost)
	return nil
}
