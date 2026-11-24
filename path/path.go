// Package path resolves per-user application paths without creating files.
package path

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ygrebnov/errorc"
)

const parentDir = ".."

// DataDir returns the OS-specific per-user data directory for appName.
// An absolute XDG_DATA_HOME is used on Unix; a relative value is ignored.
// On Windows, a non-empty APPDATA must be absolute.
func DataDir(appName string) (string, error) {
	return dataDir(appName, runtime.GOOS)
}

func dataDir(appName, goos string) (string, error) {
	if err := validateAppName(appName); err != nil {
		return "", err
	}

	switch goos {
	case "darwin":
		home, err := userHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, "Library", "Application Support", appName), nil
	case "windows":
		if base := os.Getenv("APPDATA"); base != "" {
			if !filepath.IsAbs(base) {
				return "", errorc.With(
					ErrInvalidBasePath,
					errorc.String("APPDATA", base),
				)
			}

			return filepath.Join(base, appName), nil
		}

		home, err := userHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, "AppData", "Roaming", appName), nil
	default:
		if base := os.Getenv("XDG_DATA_HOME"); base != "" && filepath.IsAbs(base) {
			return filepath.Join(base, appName), nil
		}

		home, err := userHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, ".local", "share", appName), nil
	}
}

// ConfigDir returns the OS-specific per-user config directory for appName.
// A non-empty XDG_CONFIG_HOME must be absolute.
func ConfigDir(appName string) (string, error) {
	if err := validateAppName(appName); err != nil {
		return "", err
	}

	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", errorc.WithCause(ErrCannotResolveConfig, err)
		}
	}

	if !filepath.IsAbs(base) {
		return "", errorc.With(
			ErrInvalidBasePath,
			errorc.String("XDG_CONFIG_HOME", base),
		)
	}

	return filepath.Join(base, appName), nil
}

const keyPath = "path"

// ValidateWithinHome checks lexical containment in the user's home directory.
// It does not resolve symlinks or establish a filesystem confinement boundary.
func ValidateWithinHome(path string) error {
	home, err := userHomeDir()
	if err != nil {
		return err
	}

	if !filepath.IsAbs(path) {
		return errorc.With(ErrInvalidBasePath, errorc.String(keyPath, path))
	}

	rel, err := filepath.Rel(home, path)
	if err != nil ||
		rel == parentDir ||
		filepath.IsAbs(rel) ||
		strings.HasPrefix(rel, parentDir+string(filepath.Separator)) {
		return errorc.With(ErrOutsideHome, errorc.String(keyPath, path))
	}

	return nil
}

func userHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errorc.WithCause(ErrCannotResolveHome, err)
	}

	return home, nil
}

func validateAppName(name string) error {
	if name == "" ||
		name == "." ||
		name == parentDir ||
		strings.TrimSpace(name) != name ||
		strings.ContainsAny(name, `/\`) ||
		filepath.IsAbs(name) ||
		strings.ContainsAny(name, `<>:"|?*`) ||
		strings.HasSuffix(name, ".") {
		return ErrInvalidAppName
	}

	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return ErrInvalidAppName
		}
	}

	base := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	switch base {
	case "CON", "PRN", "AUX", "NUL", "CONIN$", "CONOUT$":
		return ErrInvalidAppName
	}

	if len(base) == 4 &&
		(strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) &&
		base[3] >= '1' &&
		base[3] <= '9' {
		return ErrInvalidAppName
	}

	return nil
}
