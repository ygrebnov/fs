//go:build !windows

package fs

import "os"

func validatePrivatePermissions(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0o077 != 0 {
		return ErrInsecurePermissions
	}
	return nil
}

func setPrivateDirectory(dir *os.File) error {
	return dir.Chmod(0o700)
}
