package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func stopSandbox(id string) error {
	pidPath := filepath.Join(basePath, id, "pid")

	data, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("cannot read sandbox: %w", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid pid: %w", err)
	}
	if !isAlive(pid) {
		fmt.Printf("[forker] sandbox %s already stopped\n", id)
		return nil
	}
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("cannot kill sandbox: %w", err)
	}

	for i := 0; i < 5; i++ {
		if !isAlive(pid) {
			fmt.Printf("[forker] sandbox %s stopped\n", id)
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("[forker] forcing stop...\n")
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("cannot kill sandbox: %w", err)
	}

	fmt.Printf("[forker] sandbox %s stopped\n", id)

	_ = cleanupVeth(id)

	if err := os.RemoveAll(filepath.Join(basePath, id)); err != nil {
		fmt.Printf("[forker] warning: failed to remove sandbox state: %v\n", err)
	}

	return nil
}
