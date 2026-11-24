//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package fs

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestPrivatePermissions(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "private")
	if err := os.WriteFile(filePath, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := ValidatePrivateFile(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filePath, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePrivateFile(file); !errors.Is(err, ErrInsecurePermissions) {
		t.Fatalf("public file error = %v", err)
	}
	dirFile, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer dirFile.Close()
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePrivateDirectory(dirFile); !errors.Is(err, ErrInsecurePermissions) {
		t.Fatalf("public directory error = %v", err)
	}
	if err := SetPrivateDirectory(dirFile); err != nil {
		t.Fatal(err)
	}
}

func TestFIFORejectedWithoutBlocking(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skipf("mkfifo: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := ReadFile(path, 16)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("FIFO accepted")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FIFO read blocked")
	}
}

func TestAppendFIFORejectedWithoutBlocking(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skipf("mkfifo: %v", err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	done := make(chan error, 1)
	go func() {
		file, err := OpenRegularAppendWithinRoot(root, "fifo", 0o600)
		if err == nil {
			_ = file.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, ErrNotRegularFile) {
			t.Fatalf("FIFO error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FIFO append blocked")
	}
}
