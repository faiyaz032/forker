func setupCgroups(id string, pid int, cfg RunConfig) error {
	base := filepath.Join("/sys/fs/cgroup", "forker", id)

	if err := os.MkdirAll(base, 0755); err != nil {
		return err
	}

	// memory
	if cfg.MemoryMax != "" {
		_ = os.WriteFile(
			filepath.Join(base, "memory.max"),
			[]byte(cfg.MemoryMax),
			0644,
		)
	}

	// cpu
	cpu := cpuToCgroupQuota(cfg.CPUQuota)
	_ = os.WriteFile(
		filepath.Join(base, "cpu.max"),
		[]byte(cpu),
		0644,
	)

	// pids
	if cfg.PidsMax > 0 {
		_ = os.WriteFile(
			filepath.Join(base, "pids.max"),
			[]byte(strconv.Itoa(cfg.PidsMax)),
			0644,
		)
	}

	// attach process
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