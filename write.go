package fs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ygrebnov/errorc"
)

// WriteFileAtomic writes data to a new regular file and renames it over name.
// root must be the already opened parent directory; name must be a single
// component. Existing symlinks and non-regular files are rejected. The
// replacement is atomic, but this function does not promise crash durability.
func WriteFileAtomic(root *os.Root, name string, data []byte, perm os.FileMode) error {
	if name == "" ||
		name == "." ||
		name == ".." ||
		filepath.Base(name) != name ||
		strings.ContainsAny(name, `/\`) {
		return errorc.With(
			ErrInvalidFileName,
			errorc.String(keyValidation, "file name must be a single path component"),
			errorc.String(keyName, name),
		)
	}

	info, err := root.Lstat(name)
	switch {
	case err == nil && !info.Mode().IsRegular():
		return ErrNotRegularFile
	case err != nil && !errors.Is(err, os.ErrNotExist):
		return err
	}

	tempName, file, err := CreateTempFileWithinRoot(root, ".tmp-", "", perm)
	if err != nil {
		return err
	}
	defer func() { _ = root.Remove(tempName) }()

	if n, errWrite := file.Write(data); errWrite != nil {
		_ = file.Close()
		return errWrite
	} else if n != len(data) {
		_ = file.Close()
		return io.ErrShortWrite
	}

	if errSync := file.Sync(); errSync != nil {
		_ = file.Close()
		return errSync
	}

	if errClose := file.Close(); errClose != nil {
		return errClose
	}

	return root.Rename(tempName, name)
}
