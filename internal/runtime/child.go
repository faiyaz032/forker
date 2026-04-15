package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
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
		return fmt.Errorf("hostname failed: %w", err)
	}

	if err := setupMounts(); err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}

	if err := setupNetworking(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "[forker] network warning: %v\n", err)
	}

	pid := os.Getpid()

	if err := saveSandbox(cfg.SandboxID, pid, cfg.Bin); err != nil {
		return fmt.Errorf("save sandbox failed: %w", err)
	}

	args := append([]string{cfg.Bin}, cfg.Args...)
	return syscall.Exec(cfg.Bin, args, os.Environ())
}

func loadConfig() Config {
	var args []string

	_ = json.Unmarshal([]byte(os.Getenv("__FORKER_ARGS__")), &args)

	return Config{
		Bin:       mustEnv("__FORKER_BIN__"),
		Args:      args,
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
