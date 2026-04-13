package runtime

import (
	"fmt"
	"os"
	"os/exec"
)

type Config struct {
	Bin       string
	Args      []string
	Hostname  string
	SandboxID string
}

func ChildMain() error {
	cfg := loadConfig()

	fmt.Printf("[forker][%s] child started\n", cfg.SandboxID)

	if err := setHostname(cfg); err != nil {
		return err
	}

	if err := setupMounts(); err != nil {
		return err
	}

	_ = setupNetworking(cfg)

	serverCmd := exec.Command(cfg.Bin, cfg.Args...)
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	serverCmd.Stdin = nil

	if err := serverCmd.Start(); err != nil {
		return err
	}

	fmt.Printf("[forker][%s] server started (pid=%d)\n", cfg.SandboxID, serverCmd.Process.Pid)

	shell := exec.Command("/bin/bash")
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr
	shell.Stdin = os.Stdin

	return shell.Run()
}

func loadConfig() Config {
	return Config{
		Bin:       mustEnv("__FORKER_BIN__"),
		Args:      decodeArgs(os.Getenv("__FORKER_ARGS__")),
		Hostname:  os.Getenv("__FORKER_HOSTNAME__"),
		SandboxID: os.Getenv("__FORKER_SANDBOX_ID__"),
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(key + " not set")
	}
	return v
}
