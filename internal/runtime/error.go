package runtime

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func logSyscallError(op string, err error) {
	if err == nil {
		return
	}
	// pull out the raw syscall errno so we can log exactly what happened under the hood (e.g. eperm or enoent)
	var errno syscall.Errno
	if e, ok := err.(syscall.Errno); ok {
		errno = e
	} else if pathErr, ok := err.(*os.PathError); ok {
		if e, ok := pathErr.Err.(syscall.Errno); ok {
			errno = e
		}
	} else if syscallErr, ok := err.(*os.SyscallError); ok {
		if e, ok := syscallErr.Err.(syscall.Errno); ok {
			errno = e
		}
	} else if errors.As(err, &errno) {
	}
	if errno != 0 {
		var name string
		switch errno {
		case syscall.EPERM:
			name = "EPERM"
		case syscall.ENOENT:
			name = "ENOENT"
		case syscall.EACCES:
			name = "EACCES"
		case syscall.EINVAL:
			name = "EINVAL"
		case syscall.ENOTDIR:
			name = "ENOTDIR"
		case syscall.EISDIR:
			name = "EISDIR"
		case syscall.EBUSY:
			name = "EBUSY"
		case syscall.EEXIST:
			name = "EEXIST"
		case syscall.ENOMEM:
			name = "ENOMEM"
		default:
			name = fmt.Sprintf("ERRNO_%d", errno)
		}
		fmt.Fprintf(os.Stderr, "[forker] syscall error during %s: %s (%s)\n", op, name, errno.Error())
	} else {
		fmt.Fprintf(os.Stderr, "[forker] error during %s: %v\n", op, err)
	}
}
