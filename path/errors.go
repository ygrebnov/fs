package path

import "errors"

var (
	ErrInvalidAppName      = errors.New("invalid application name")
	ErrInvalidBasePath     = errors.New("invalid application base path")
	ErrOutsideHome         = errors.New("application path is outside user home")
	ErrCannotResolveHome   = errors.New("cannot resolve user home directory")
	ErrCannotResolveConfig = errors.New("cannot resolve user config directory")
)
