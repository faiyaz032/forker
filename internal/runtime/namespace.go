package runtime

import "syscall"

func setHostname(cfg Config) error {
	err := syscall.Sethostname([]byte(cfg.Hostname))
	if err != nil {
		logSyscallError("setHostname", err)
	}
	return err
}
