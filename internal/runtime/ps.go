package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func listSandboxes() error {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	fmt.Printf("%-20s %s\n", "ID", "STATUS")

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		id := e.Name()

		pidPath := filepath.Join(basePath, id, "pid")

		pidBytes, err := os.ReadFile(pidPath)
		if err != nil {
			fmt.Printf("[ps] cannot read pid file for %s: %v\n", id, err)
			continue
		}

		pidStr := strings.TrimSpace(string(pidBytes))

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			fmt.Printf("[ps] invalid pid for %s: %v\n", id, err)
			continue
		}

		status := "Running"
		if !isAlive(pid) {
			status = "Stopped"
		}

		fmt.Printf("%-20s %s\n", id, status)
	}

	return nil
}

func isAlive(pid int) bool {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	return err == nil && len(data) > 0
}
