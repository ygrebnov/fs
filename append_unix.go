//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package fs

import (
	"os"
	"syscall"
)

func openRootAppendFile(root *os.Root, name string, perm os.FileMode) (*os.File, error) {
	return root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND|syscall.O_NONBLOCK, perm)
}
