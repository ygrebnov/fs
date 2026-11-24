package fs

import (
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ygrebnov/errorc"
)

func TestOpenRegularRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input")
	if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}

	file, err := OpenRegularRead(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if string(data) != "content" {
		t.Fatalf("data = %q", data)
	}
	if _, err := OpenRegularRead(dir); !errors.Is(err, ErrNotRegularFile) {
		t.Fatalf("directory error = %v", err)
	}
	if _, err := OpenRegularRead(filepath.Join(dir, "missing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file error = %v", err)
	}
}

func TestOpenRegularReadWithinRoot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "inside"), []byte("safe"), 0o600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	file, err := OpenRegularReadWithinRoot(root, "inside")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if string(data) != "safe" {
		t.Fatalf("data = %q", data)
	}
	if _, err := OpenRegularReadWithinRoot(root, "."); !errors.Is(err, ErrNotRegularFile) {
		t.Fatalf("directory error = %v", err)
	}
	if _, err := OpenRegularReadWithinRoot(root, "missing"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file error = %v", err)
	}
	if _, err := OpenRegularReadWithinRoot(root, "../outside"); err == nil {
		t.Fatal("path traversal accepted")
	}
}

func TestOpenRegularAppendWithinRoot(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	for _, data := range []string{"first", "second"} {
		file, err := OpenRegularAppendWithinRoot(root, "app.log", 0o600)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.WriteString(data); err != nil {
			_ = file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	content, err := os.ReadFile(filepath.Join(dir, "app.log"))
	if err != nil || string(content) != "firstsecond" {
		t.Fatalf("content = %q, %v", content, err)
	}
	if _, err := OpenRegularAppendWithinRoot(root, ".", 0o600); !errors.Is(err, ErrNotRegularFile) {
		t.Fatalf("directory error = %v", err)
	}
	if _, err := OpenRegularAppendWithinRoot(root, "../outside", 0o600); err == nil {
		t.Fatal("path traversal accepted")
	}
}

func TestValidateRegularFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateRegularFile(file); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRegularFile(file); err == nil {
		t.Fatal("closed file accepted")
	}
	directory, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer directory.Close()
	if err := ValidateRegularFile(directory); !errors.Is(err, ErrNotRegularFile) {
		t.Fatalf("directory error = %v", err)
	}
}

func TestPrivatePermissionTypeErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	directory, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer directory.Close()

	if err := ValidatePrivateFile(directory); !errors.Is(err, ErrNotRegularFile) {
		t.Fatalf("private file type error = %v", err)
	}
	if err := ValidatePrivateDirectory(file); !errors.Is(err, ErrNotDirectory) {
		t.Fatalf("private directory type error = %v", err)
	}
	if err := SetPrivateDirectory(file); !errors.Is(err, ErrNotDirectory) {
		t.Fatalf("set private directory type error = %v", err)
	}
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input")
	if err := os.WriteFile(path, []byte("1234"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		limit int64
		want  error
	}{
		{"exact", 4, nil},
		{"too small", 3, ErrExceedsFileSizeLimit},
		{"invalid", -1, ErrInvalidFileSizeLimit},
		{"overflow", math.MaxInt64, ErrInvalidFileSizeLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := ReadFile(path, tc.limit)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			if tc.want == nil && string(data) != "1234" {
				t.Fatalf("data = %q", data)
			}
			if tc.want != nil {
				if limit, ok := errorc.Lookup[int64](err, keyLimit); !ok || limit != tc.limit {
					t.Fatalf("limit field = %d, %v", limit, ok)
				}
			}
		})
	}
	if _, err := ReadFile(dir, 100); !errors.Is(err, ErrNotRegularFile) {
		t.Fatalf("directory error = %v", err)
	}
}

func TestReadFileWithinRoot(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "inside"), []byte("safe"), 0o600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	data, err := ReadFileWithinRoot(root, "inside", 4)
	if err != nil || string(data) != "safe" {
		t.Fatalf("read = %q, %v", data, err)
	}
	if _, err := ReadFileWithinRoot(root, outside, 100); err == nil {
		t.Fatal("absolute path accepted")
	}
	if err := os.Symlink(outside, filepath.Join(dir, "escape")); err == nil {
		if _, err := ReadFileWithinRoot(root, "escape", 100); err == nil {
			t.Fatal("escaping symlink accepted")
		}
	}
}

func TestReadOpenedRegularFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input")
	if err := os.WriteFile(path, []byte("1234"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		limit int64
		want  error
	}{
		{"exact", 4, nil},
		{"too small", 3, ErrExceedsFileSizeLimit},
		{"invalid", -1, ErrInvalidFileSizeLimit},
		{"overflow", math.MaxInt64, ErrInvalidFileSizeLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, err := OpenRegularRead(path)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			data, err := ReadOpenedRegularFile(file, tc.limit)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			if tc.want == nil && string(data) != "1234" {
				t.Fatalf("data = %q", data)
			}
			if tc.want != nil {
				if limit, ok := errorc.Lookup[int64](err, keyLimit); !ok || limit != tc.limit {
					t.Fatalf("limit field = %d, %v", limit, ok)
				}
			}
		})
	}
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	defer directory.Close()
	if _, err := ReadOpenedRegularFile(directory, 4); !errors.Is(err, ErrNotRegularFile) {
		t.Fatalf("directory error = %v", err)
	}
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := WriteFileAtomic(root, "config.json", []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(root, "config.json", []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil || string(data) != "second" {
		t.Fatalf("replacement = %q, %v", data, err)
	}
	if err := WriteFileAtomic(root, "../escape", nil, 0o600); !errors.Is(err, ErrInvalidFileName) {
		t.Fatalf("traversal error = %v", err)
	} else {
		if name, ok := errorc.Lookup[string](err, keyName); !ok || name != "../escape" {
			t.Fatalf("name field = %q, %v", name, ok)
		}
		if validation, ok := errorc.Lookup[string](err, keyValidation); !ok || validation == "" {
			t.Fatalf("validation field = %q, %v", validation, ok)
		}
	}
	if err := WriteFileAtomic(root, ".", nil, 0o600); !errors.Is(err, ErrInvalidFileName) {
		t.Fatalf("directory name error = %v", err)
	}
	if err := os.Symlink("config.json", filepath.Join(dir, "link")); err == nil {
		if err := WriteFileAtomic(root, "link", nil, 0o600); !errors.Is(err, ErrNotRegularFile) {
			t.Fatalf("symlink error = %v", err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if len(entry.Name()) >= 5 && entry.Name()[:5] == ".tmp-" {
			t.Fatalf("temporary file left behind: %s", entry.Name())
		}
	}
}

func TestCreateTempFileWithinRoot(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	name, file, err := CreateTempFileWithinRoot(root, "config-", ".yaml", 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(name, "config-") || !strings.HasSuffix(name, ".yaml") {
		t.Fatalf("temporary name = %q", name)
	}
	if err := ValidateRegularFile(file); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := root.Remove(name); err != nil {
		t.Fatal(err)
	}
	if _, _, err := CreateTempFileWithinRoot(root, "../", ".yaml", 0o600); !errors.Is(err, ErrInvalidFileName) {
		t.Fatalf("invalid prefix error = %v", err)
	}
	if _, _, err := CreateTempFileWithinRoot(root, "config-", "\\bad", 0o600); !errors.Is(err, ErrInvalidFileName) {
		t.Fatalf("invalid suffix error = %v", err)
	}
}
