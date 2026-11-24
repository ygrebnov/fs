//go:build windows

package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrivatePermissionsWindows(t *testing.T) {
	dir := t.TempDir()
	directory, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer directory.Close()
	if err := SetPrivateDirectory(directory); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePrivateDirectory(directory); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "private")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := ValidatePrivateFile(file); err != nil {
		t.Fatal(err)
	}
}
