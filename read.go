package fs

import (
	"io"
	"math"
	"os"

	"github.com/ygrebnov/errorc"
)

// OpenRegularRead opens a regular file for reading. The caller must close it.
// The path is chosen by the caller and is not confined to a directory.
func OpenRegularRead(path string) (*os.File, error) {
	file, err := openReadFile(path)
	if err != nil {
		return nil, err
	}

	if errValidate := ValidateRegularFile(file); errValidate != nil {
		_ = file.Close()
		return nil, errValidate
	}
	return file, nil
}

// OpenRegularReadWithinRoot opens a regular file beneath an established root.
// The caller must close it. os.Root rejects paths that escape the root.
func OpenRegularReadWithinRoot(root *os.Root, name string) (*os.File, error) {
	file, err := openRootReadFile(root, name)
	if err != nil {
		return nil, err
	}

	if errValidate := ValidateRegularFile(file); errValidate != nil {
		_ = file.Close()
		return nil, errValidate
	}
	return file, nil
}

// ValidateRegularFile checks the already opened file, avoiding a separate
// path lookup that could refer to a different file.
func ValidateRegularFile(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return ErrNotRegularFile
	}
	return nil
}

// ReadFile reads a regular file chosen by the caller, up to maxSize bytes.
// The size limit is checked before and during the read.
func ReadFile(path string, maxSize int64) ([]byte, error) {
	file, err := OpenRegularRead(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return readBounded(file, maxSize)
}

// ReadFileWithinRoot reads a regular file beneath an established root,
// up to maxSize bytes. The caller owns the root.
func ReadFileWithinRoot(root *os.Root, name string, maxSize int64) ([]byte, error) {
	file, err := OpenRegularReadWithinRoot(root, name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return readBounded(file, maxSize)
}

// ReadOpenedRegularFile reads an already opened regular file up to maxSize bytes.
// The caller retains ownership of the file. The size is checked before and
// during the read, so a growing file cannot exceed the limit.
func ReadOpenedRegularFile(file *os.File, maxSize int64) ([]byte, error) {
	if err := ValidateRegularFile(file); err != nil {
		return nil, err
	}
	return readBounded(file, maxSize)
}

func readBounded(file *os.File, maxSize int64) ([]byte, error) {
	if maxSize < 0 || maxSize == math.MaxInt64 {
		return nil, errorc.With(
			ErrInvalidFileSizeLimit,
			errorc.Int64(keyLimit, maxSize),
		)
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if info.Size() > maxSize {
		return nil, errorc.With(
			ErrExceedsFileSizeLimit,
			errorc.Int64(keyLimit, maxSize),
		)
	}

	data, err := io.ReadAll(io.LimitReader(file, maxSize+1))
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > maxSize {
		return nil, errorc.With(
			ErrExceedsFileSizeLimit,
			errorc.Int64(keyLimit, maxSize),
		)
	}

	return data, nil
}
