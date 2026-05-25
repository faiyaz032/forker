package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func setupCgroups(id string, pid int, cfg RunConfig) error {
	// enable controllers in root and parent cgroups so child directories can manage resources under cgroups v2
	_ = os.WriteFile("/sys/fs/cgroup/cgroup.subtree_control", []byte("+cpu +memory +pids"), 0644)
	parent := filepath.Join("/sys/fs/cgroup", "forker")
	_ = os.MkdirAll(parent, 0755)
	_ = os.WriteFile(filepath.Join(parent, "cgroup.subtree_control"), []byte("+cpu +memory +pids"), 0644)

	base := filepath.Join(parent, id)
	if err := os.MkdirAll(base, 0755); err != nil {
		return err
	}

	if cfg.MemoryMax != "" {
		_ = os.WriteFile(
			filepath.Join(base, "memory.max"),
			[]byte(cfg.MemoryMax),
			0644,
		)
	}

	cpu := cpuToCgroupQuota(cfg.CPUQuota)
	_ = os.WriteFile(
		filepath.Join(base, "cpu.max"),
		[]byte(cpu),
		0644,
	)

	if cfg.PidsMax > 0 {
		_ = os.WriteFile(
			filepath.Join(base, "pids.max"),
			[]byte(strconv.Itoa(cfg.PidsMax)),
			0644,
		)
	}

	return os.WriteFile(
		filepath.Join(base, "cgroup.procs"),
		[]byte(strconv.Itoa(pid)),
		0644,
	)
}

func cpuToCgroupQuota(cpu float64) string {
	if cpu <= 0 {
		return "max"
	}

	period := 100000
	quota := int64(float64(period) * cpu)

	return fmt.Sprintf("%d %d", quota, period)
}
