//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package fs

import "os"

// The precheck avoids opening a known special file on platforms without
// O_NONBLOCK. It cannot prevent a replacement between Stat and Open.
func openReadFile(path string) (*os.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, ErrNotRegularFile
	}

	return os.Open(path)
}

func openRootReadFile(root *os.Root, name string) (*os.File, error) {
	info, err := root.Stat(name)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, ErrNotRegularFile
	}

	return root.Open(name)
}
