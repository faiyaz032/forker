package runtime

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestLogSyscallErrorNilIsSilent(t *testing.T) {
	_, stderr := captureOutput(t, func() {
		logSyscallError("noop", nil)
	})

	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestLogSyscallErrorKnownErrnoNames(t *testing.T) {
	tests := []struct {
		errno syscall.Errno
		want  string
	}{
		{syscall.EPERM, "EPERM"},
		{syscall.ENOENT, "ENOENT"},
		{syscall.EACCES, "EACCES"},
		{syscall.EINVAL, "EINVAL"},
		{syscall.ENOTDIR, "ENOTDIR"},
		{syscall.EISDIR, "EISDIR"},
		{syscall.EBUSY, "EBUSY"},
		{syscall.EEXIST, "EEXIST"},
		{syscall.ENOMEM, "ENOMEM"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			_, stderr := captureOutput(t, func() {
				logSyscallError("mount", tt.errno)
			})

			if !strings.Contains(stderr, "syscall error during mount") {
				t.Errorf("stderr = %q, want operation name", stderr)
			}
			if !strings.Contains(stderr, "("+tt.want+")") && !strings.Contains(stderr, ": "+tt.want+" ") {
				t.Errorf("stderr = %q, want errno name %q", stderr, tt.want)
			}
			if !strings.Contains(stderr, tt.errno.Error()) {
				t.Errorf("stderr = %q, want errno text %q", stderr, tt.errno.Error())
			}
		})
	}
}

func TestLogSyscallErrorUnknownErrno(t *testing.T) {
	errno := syscall.Errno(0x7f)

	_, stderr := captureOutput(t, func() {
		logSyscallError("weird", errno)
	})

	if !strings.Contains(stderr, "ERRNO_127") {
		t.Errorf("stderr = %q, want %q", stderr, "ERRNO_127")
	}
}

func TestLogSyscallErrorPathError(t *testing.T) {
	err := &os.PathError{Op: "open", Path: "/nope", Err: syscall.ENOENT}

	_, stderr := captureOutput(t, func() {
		logSyscallError("open", err)
	})

	if !strings.Contains(stderr, "ENOENT") {
		t.Errorf("stderr = %q, want ENOENT extracted from PathError", stderr)
	}
	if !strings.Contains(stderr, "syscall error during open") {
		t.Errorf("stderr = %q, want syscall error prefix", stderr)
	}
}

func TestLogSyscallErrorSyscallError(t *testing.T) {
	err := os.NewSyscallError("setns", syscall.EACCES)

	_, stderr := captureOutput(t, func() {
		logSyscallError("setns", err)
	})

	if !strings.Contains(stderr, "EACCES") {
		t.Errorf("stderr = %q, want EACCES extracted from SyscallError", stderr)
	}
}

func TestLogSyscallErrorWrappedErrno(t *testing.T) {
	err := fmt.Errorf("mount proc: %w", syscall.EPERM)

	_, stderr := captureOutput(t, func() {
		logSyscallError("mount", err)
	})

	if !strings.Contains(stderr, "EPERM") {
		t.Errorf("stderr = %q, want EPERM unwrapped", stderr)
	}
}

func TestLogSyscallErrorDeeplyWrappedErrno(t *testing.T) {
	err := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", syscall.EBUSY))

	_, stderr := captureOutput(t, func() {
		logSyscallError("umount", err)
	})

	if !strings.Contains(stderr, "EBUSY") {
		t.Errorf("stderr = %q, want EBUSY unwrapped", stderr)
	}
}

func TestLogSyscallErrorPlainError(t *testing.T) {
	err := errors.New("something broke")

	_, stderr := captureOutput(t, func() {
		logSyscallError("start", err)
	})

	if !strings.Contains(stderr, "error during start: something broke") {
		t.Errorf("stderr = %q, want plain error format", stderr)
	}
	if strings.Contains(stderr, "syscall error") {
		t.Errorf("stderr = %q, want no syscall wording for plain error", stderr)
	}
}

func TestLogSyscallErrorZeroErrnoUsesPlainFormat(t *testing.T) {
	_, stderr := captureOutput(t, func() {
		logSyscallError("zero", syscall.Errno(0))
	})

	if !strings.Contains(stderr, "error during zero") {
		t.Errorf("stderr = %q, want plain error format", stderr)
	}
	if strings.Contains(stderr, "syscall error") {
		t.Errorf("stderr = %q, want no syscall wording for zero errno", stderr)
	}
}

func TestLogSyscallErrorPathErrorWithoutErrno(t *testing.T) {
	err := &os.PathError{Op: "open", Path: "/nope", Err: errors.New("custom")}

	_, stderr := captureOutput(t, func() {
		logSyscallError("open", err)
	})

	if !strings.Contains(stderr, "error during open") {
		t.Errorf("stderr = %q, want plain error format", stderr)
	}
}
