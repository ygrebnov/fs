package fs

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

// CreateTempFileWithinRoot creates a new file beneath root using a random
// single-component name. The caller owns the file and must close and remove it.
// Prefix and suffix must not contain path separators.
func CreateTempFileWithinRoot(root *os.Root, prefix, suffix string, perm os.FileMode) (string, *os.File, error) {
	if strings.ContainsAny(prefix, `/\`) || strings.ContainsAny(suffix, `/\`) {
		return "", nil, ErrInvalidFileName
	}

	for range 10 {
		random := make([]byte, 12)
		if _, err := rand.Read(random); err != nil {
			return "", nil, err
		}
		name := prefix + hex.EncodeToString(random) + suffix
		file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", nil, err
		}
		return name, file, nil
	}
	return "", nil, os.ErrExist
}
