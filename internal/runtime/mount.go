package runtime

import (
	"fmt"
	"syscall"
)

func setupMounts() error {
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		logSyscallError("mount private /", err)
		return err
	}

	if err := syscall.Unmount("/proc", syscall.MNT_DETACH); err != nil {
		logSyscallError("unmount /proc", err)
	}

	if err := syscall.Mount("proc", "/proc", "proc",
		syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC, ""); err != nil {
		logSyscallError("mount proc", err)
		return fmt.Errorf("mount proc: %w", err)
	}

	if err := syscall.Mount("tmpfs", "/tmp", "tmpfs",
		syscall.MS_NOSUID|syscall.MS_NODEV, "size=64m"); err != nil {
		logSyscallError("mount tmpfs /tmp", err)
		return fmt.Errorf("mount tmp: %w", err)
	}

	return nil
}
