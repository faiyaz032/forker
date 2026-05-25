package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type Config struct {
	Bin       string
	Args      []string
	Hostname  string
	SandboxID string
}

func ChildMain() (err error) {
	cfg := loadConfig()

	fmt.Printf("[forker][%s] child started\n", cfg.SandboxID)

	defer func() {
		// if things go sideways, make sure we clean up the mounts to avoid leaving trace on the host
		if r := recover(); r != nil {
			_ = syscall.Unmount("/proc", syscall.MNT_DETACH)
			_ = syscall.Unmount("/tmp", syscall.MNT_DETACH)
			panic(r)
		} else if err != nil {
			_ = syscall.Unmount("/proc", syscall.MNT_DETACH)
			_ = syscall.Unmount("/tmp", syscall.MNT_DETACH)
		}
	}()

	if err = setHostname(cfg); err != nil {
		return err
	}

	if err = setupMounts(); err != nil {
		return err
	}

	if err = setupNetworking(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "[forker] network warning: %v\n", err)
	}

	readyFile := fmt.Sprintf("/var/run/forker/%s/ready", cfg.SandboxID)
	if err = os.WriteFile(readyFile, []byte("1"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "[forker] warning: failed to write readiness file: %v\n", err)
	}

	args := append([]string{cfg.Bin}, cfg.Args...)
	err = syscall.Exec(cfg.Bin, args, os.Environ())
	if err != nil {
		logSyscallError("exec", err)
	}
	return err
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

func waitForSandboxReady(id string) error {
	readyPath := filepath.Join(basePath, id, "ready")

	for i := 0; i < 50; i++ { // ~5 seconds
		if _, err := os.Stat(readyPath); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("sandbox not ready")
}
