package runtime

import (
	"fmt"
	"syscall"
)

func setupMounts() error {

	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return err
	}

	_ = syscall.Unmount("/proc", syscall.MNT_DETACH)

	if err := syscall.Mount("proc", "/proc", "proc",
		syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC, ""); err != nil {
		return fmt.Errorf("mount proc: %w", err)
	}

	_ = syscall.Mount("tmpfs", "/tmp", "tmpfs",
		syscall.MS_NOSUID|syscall.MS_NODEV, "size=64m")

	return nil
}
