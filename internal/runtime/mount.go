package runtime

import "syscall"

func setupMounts() error {
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return err
	}

	if err := syscall.Mount("", "/", "", syscall.MS_SLAVE|syscall.MS_REC, ""); err != nil {
		return err
	}

	_ = syscall.Mount("proc", "/proc", "proc",
		syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC, "")

	_ = syscall.Mount("tmpfs", "/tmp", "tmpfs",
		syscall.MS_NOSUID|syscall.MS_NODEV, "size=64m")

	return nil
}
