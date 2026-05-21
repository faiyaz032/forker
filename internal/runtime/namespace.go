package runtime

import "syscall"

func setHostname(cfg Config) error {
	return syscall.Sethostname([]byte(cfg.Hostname))
}
