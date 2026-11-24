//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package fs

import (
	"os"
	"syscall"
)

// O_NONBLOCK keeps a FIFO or device from blocking before Stat can reject it.
func openReadFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}

func openRootReadFile(root *os.Root, name string) (*os.File, error) {
	return root.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
