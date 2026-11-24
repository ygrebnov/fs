package fs_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ygrebnov/fs"
	fspath "github.com/ygrebnov/fs/path"
)

// A config loader checks permissions on the same handle it later reads, then
// replaces the file atomically when saving an updated configuration.
func Example_privateConfigFile() {
	dir, err := os.MkdirTemp("", "fs-config-example-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	directory, err := os.Open(dir)
	if err != nil {
		log.Fatal(err)
	}
	if err := fs.SetPrivateDirectory(directory); err != nil {
		log.Fatal(err)
	}
	if err := directory.Close(); err != nil {
		log.Fatal(err)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		log.Fatal(err)
	}
	defer root.Close()

	if err := fs.WriteFileAtomic(root, "config.yml", []byte("name: first\n"), 0o600); err != nil {
		log.Fatal(err)
	}
	file, err := fs.OpenRegularReadWithinRoot(root, "config.yml")
	if err != nil {
		log.Fatal(err)
	}
	if err := fs.ValidatePrivateFile(file); err != nil {
		log.Fatal(err)
	}
	loaded, err := fs.ReadOpenedRegularFile(file, 1<<20)
	if err != nil {
		log.Fatal(err)
	}
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}

	if err := fs.WriteFileAtomic(root, "config.yml", []byte("name: second\n"), 0o600); err != nil {
		log.Fatal(err)
	}
	updated, err := fs.ReadFileWithinRoot(root, "config.yml", 1<<20)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(loaded), string(updated))
	// Output:
	// name: first
	// name: second
}

// A process resolves its data and config directories independently,
// then applies its home-boundary policy before creating files in either one.
func Example_applicationPaths() {
	home, err := os.MkdirTemp("", "fs-path-example-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(home) }()

	for name, value := range map[string]string{
		"HOME":            home,
		"USERPROFILE":     home,
		"APPDATA":         filepath.Join(home, "AppData", "Roaming"),
		"XDG_DATA_HOME":   filepath.Join(home, ".local", "share"),
		"XDG_CONFIG_HOME": filepath.Join(home, ".config"),
	} {
		previous, existed := os.LookupEnv(name)
		if err := os.Setenv(name, value); err != nil {
			log.Fatal(err)
		}
		defer func() {
			if existed {
				_ = os.Setenv(name, previous)
			} else {
				_ = os.Unsetenv(name)
			}
		}()
	}

	dataDir, err := fspath.DataDir("e2engine")
	if err != nil {
		log.Fatal(err)
	}
	configDir, err := fspath.ConfigDir("e2engine")
	if err != nil {
		log.Fatal(err)
	}
	if err := fspath.ValidateWithinHome(dataDir); err != nil {
		log.Fatal(err)
	}
	if err := fspath.ValidateWithinHome(configDir); err != nil {
		log.Fatal(err)
	}
	fmt.Println(filepath.Base(dataDir), filepath.Base(configDir))
	// Output: e2engine e2engine
}
