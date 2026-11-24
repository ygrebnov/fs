**fs** provides reusable filesystem primitives for secure reads and writes. Child **path** package resolves per-user application paths.

[![GoDoc](https://pkg.go.dev/badge/github.com/ygrebnov/fs)](https://pkg.go.dev/github.com/ygrebnov/fs)
[![Build Status](https://github.com/ygrebnov/fs/actions/workflows/build.yml/badge.svg)](https://github.com/ygrebnov/fs/actions/workflows/build.yml)

# Usage

See [example_test.go](example_test.go) for a private config-file workflow and
per-user path resolution examples.

## Choosing an API

- Use an `os.Root` variant when a file name must stay beneath a trusted
  directory. Open the root first and pass a relative name.
- Use `ReadFile` or `OpenRegularRead` when the caller has intentionally chosen
  an unrestricted path. Both reject non-regular files.
- Keep a file handle open between permission validation and reading when both
  checks must apply to the same file. Use `OpenRegularReadWithinRoot`, then
  `ValidatePrivateFile`, then `ReadOpenedRegularFile`.
- Use `OpenRegularAppendWithinRoot` for logs and `WriteFileAtomic` for replacing
  a config-like file. The latter is an atomic replacement, not a promise of
  crash durability.
- Use `fs/path` for default per-user locations. `ValidateWithinHome` checks
  lexical containment; use rooted filesystem operations for confinement.

The caller owns handles returned by open functions and must close them. The
caller also owns every `os.Root` passed to this package.

## Reading regular files

| API | Use |
| --- | --- |
| `OpenRegularRead(path)` | Open a caller-chosen regular file and keep its handle. |
| `OpenRegularReadWithinRoot(root, name)` | Open a regular file beneath a root. |
| `ReadFile(path, maxSize)` | Read a caller-chosen regular file with a byte limit. |
| `ReadFileWithinRoot(root, name, maxSize)` | Read beneath a root with a byte limit. |
| `ReadOpenedRegularFile(file, maxSize)` | Bound a read on an existing handle, after other handle-based checks. |
| `ValidateRegularFile(file)` | Check the type of an already opened file. |

Bounded reads check size before and during reading, so a growing file cannot
exceed `maxSize`. The limit may be zero for an empty file; negative values and
`math.MaxInt64` are invalid.

## Private files and directories

| API | Use |
| --- | --- |
| `ValidatePrivateFile(file)` | Require a regular file with private permissions. |
| `ValidatePrivateDirectory(dir)` | Require a directory with private permissions. |
| `SetPrivateDirectory(dir)` | Restrict an opened directory, then verify it. |

Unix checks reject group or other access. Windows checks use ACLs rather than
Unix permission bits.

## Writing files

| API | Use |
| --- | --- |
| `OpenRegularAppendWithinRoot(root, name, perm)` | Open or create a regular file for append. |
| `CreateTempFileWithinRoot(root, prefix, suffix, perm)` | Create a named temporary file beneath a root. |
| `WriteFileAtomic(root, name, data, perm)` | Write a new file and rename it over a single-component destination. |

`WriteFileAtomic` rejects an existing symlink or non-regular destination.
SQLite databases and other files with companion files or their own locking
protocol need an application-specific open strategy.

## Per-user application paths

Import `github.com/ygrebnov/fs/path` for these functions:

| API | Use |
| --- | --- |
| `DataDir(appName)` | Resolve the per-user application data directory. |
| `ConfigDir(appName)` | Resolve the per-user application config directory. |
| `ValidateWithinHome(path)` | Check that a resolved path is lexically beneath the user's home. |

Both resolvers validate `appName` as a single path component, including Windows
reserved names. On Unix, a relative `XDG_DATA_HOME` is ignored in favor of the
home-based default; a relative `XDG_CONFIG_HOME` is rejected. On Windows, a
non-empty relative `APPDATA` is rejected. Absolute data-home overrides may be
outside the user's home; applications that require a home boundary should call
`ValidateWithinHome` before using the result.

These functions only resolve names. The application decides whether to create
the directory, whether to require private permissions, and which filename to
use inside it.

## Installation

Compatible with Go 1.26.11 or later:

```shell
go get github.com/ygrebnov/fs
```

## Versioning

This library is pre-1.0; minor version bumps may include breaking changes.

## Contributing

Contributions are welcome. Open an [issue](https://github.com/ygrebnov/fs/issues) or submit a [pull request](https://github.com/ygrebnov/fs/pulls).

## License

Distributed under the BSD 3-Clause License. See [LICENSE](LICENSE).
