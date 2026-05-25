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

	if isAlive(pid) {
		if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
			logSyscallError("kill SIGTERM", err)
			return fmt.Errorf("cannot kill sandbox: %w", err)
		}

		stopped := false
		for i := 0; i < 5; i++ {
			if !isAlive(pid) {
				stopped = true
				break
			}
			time.Sleep(200 * time.Millisecond)
		}

		if !stopped {
			fmt.Printf("[forker] forcing stop...\n")
			if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
				logSyscallError("kill SIGKILL", err)
				return fmt.Errorf("cannot kill sandbox: %w", err)
			}
		}
	}

	fmt.Printf("[forker] sandbox %s stopped\n", id)

	// clean up network, cgroups, and state files regardless of whether the sandbox died nicely or had to be killed
	_ = cleanupVeth(id)

	cgroupPath := filepath.Join("/sys/fs/cgroup", "forker", id)
	_ = os.Remove(cgroupPath)

	if err := os.RemoveAll(filepath.Join(basePath, id)); err != nil {
		fmt.Printf("[forker] warning: failed to remove sandbox state: %v\n", err)
	}

	return nil
}
