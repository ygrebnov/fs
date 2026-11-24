package path

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAppNameValidation(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../escape", `nested\app`, " app", "app.", "CON", "COM1.txt", "bad:name", "bad\x00name"} {
		if _, err := DataDir(name); !errors.Is(err, ErrInvalidAppName) {
			t.Errorf("DataDir(%q) error = %v", name, err)
		}
		if _, err := ConfigDir(name); !errors.Is(err, ErrInvalidAppName) {
			t.Errorf("ConfigDir(%q) error = %v", name, err)
		}
	}
}

func TestConfigDirAndWithinHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	base := filepath.Join(home, ".config")
	t.Setenv("XDG_CONFIG_HOME", base)
	dir, err := ConfigDir("myapp")
	if err != nil || dir != filepath.Join(base, "myapp") {
		t.Fatalf("ConfigDir() = %q, %v", dir, err)
	}
	if err := ValidateWithinHome(dir); err != nil {
		t.Fatalf("ValidateWithinHome(%q): %v", dir, err)
	}
	if err := ValidateWithinHome(filepath.Dir(home)); !errors.Is(err, ErrOutsideHome) {
		t.Fatalf("outside-home error = %v", err)
	}
	if err := ValidateWithinHome("relative"); !errors.Is(err, ErrInvalidBasePath) {
		t.Fatalf("relative-path error = %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", "relative")
	if _, err := ConfigDir("myapp"); !errors.Is(err, ErrInvalidBasePath) {
		t.Fatalf("relative config base error = %v", err)
	}
}

func TestDataDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	var want string
	switch runtime.GOOS {
	case "darwin":
		want = filepath.Join(home, "Library", "Application Support", "myapp")
	case "windows":
		t.Setenv("APPDATA", "")
		want = filepath.Join(home, "AppData", "Roaming", "myapp")
	default:
		t.Setenv("XDG_DATA_HOME", "relative")
		want = filepath.Join(home, ".local", "share", "myapp")
	}
	dir, err := DataDir("myapp")
	if err != nil || dir != want {
		t.Fatalf("DataDir() = %q, %v; want %q", dir, err, want)
	}
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", "relative")
		if _, err := DataDir("myapp"); !errors.Is(err, ErrInvalidBasePath) {
			t.Fatalf("relative APPDATA error = %v", err)
		}
	}
	if runtime.GOOS != "darwin" {
		base := filepath.Join(t.TempDir(), "external")
		if runtime.GOOS == "windows" {
			t.Setenv("APPDATA", base)
		} else {
			t.Setenv("XDG_DATA_HOME", base)
		}
		dir, err := DataDir("myapp")
		if err != nil || dir != filepath.Join(base, "myapp") {
			t.Fatalf("external data base = %q, %v", dir, err)
		}
	}
	if _, err := os.Stat(home); err != nil {
		t.Fatal(err)
	}
}

func TestDataDirRejectsRelativeAPPDATA(t *testing.T) {
	t.Setenv("APPDATA", "relative")
	if _, err := dataDir("myapp", "windows"); !errors.Is(err, ErrInvalidBasePath) {
		t.Fatalf("relative APPDATA error = %v", err)
	}
}
