package fs

import "errors"

var (
	ErrInvalidFileName      = errors.New("invalid file name")
	ErrNotRegularFile       = errors.New("not a regular file")
	ErrExceedsFileSizeLimit = errors.New("file exceeds size limit")
	ErrInvalidFileSizeLimit = errors.New("invalid file size limit")
	ErrInsecurePermissions  = errors.New("insecure filesystem permissions")
	ErrNotDirectory         = errors.New("not a directory")
	// ErrDirectoryChanged means a directory reopened for an ACL update was not
	// the same directory as the caller's open handle. It can occur on Windows.
	ErrDirectoryChanged = errors.New("directory changed while reopening for ACL update")
)
