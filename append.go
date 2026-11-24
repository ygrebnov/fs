package fs

import (
	"errors"
	"os"
)

// OpenRegularAppendWithinRoot opens or creates a regular file for append
// beneath root. The caller must close the returned file. On Unix, the open
// is nonblocking so a substituted FIFO cannot hang before type validation.
// On other platforms, the pre-open type check cannot close that race.
func OpenRegularAppendWithinRoot(root *os.Root, name string, perm os.FileMode) (*os.File, error) {
	if info, err := root.Stat(name); err == nil {
		if !info.Mode().IsRegular() {
			return nil, ErrNotRegularFile
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	file, err := openRootAppendFile(root, name, perm)
	if err != nil {
		return nil, err
	}
	if err := ValidateRegularFile(file); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}
