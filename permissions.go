package fs

import (
	"os"
)

// ValidatePrivateFile checks the permissions of an already opened file.
// On Unix, group and other permissions are forbidden. On Windows, only the
// current user, Local System, and Administrators may have allow entries.
func ValidatePrivateFile(file *os.File) error {
	if err := ValidateRegularFile(file); err != nil {
		return err
	}
	return validatePrivatePermissions(file)
}

// ValidatePrivateDirectory checks an already opened directory.
func ValidatePrivateDirectory(dir *os.File) error {
	info, err := dir.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return ErrNotDirectory
	}

	return validatePrivatePermissions(dir)
}

// SetPrivateDirectory restricts an already opened directory to the current
// user (and required system principals on Windows), then verifies the result.
func SetPrivateDirectory(dir *os.File) error {
	info, err := dir.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return ErrNotDirectory
	}

	if errSet := setPrivateDirectory(dir); errSet != nil {
		return errSet
	}

	return ValidatePrivateDirectory(dir)
}
