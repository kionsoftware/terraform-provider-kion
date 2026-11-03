// Package kfs is the filesystem seam for the kgen generators. Generators write
// through an FS so their behavior (which files get written, with what content,
// skip-vs-overwrite, mkdir failures) is unit-testable against a mock instead of
// touching the real disk.
package kfs

import (
	"io/fs"
	"os"
)

// FS is the minimal filesystem surface the generators need.
type FS interface {
	// Stat reports whether a path exists (used for skip-if-exists and for
	// locating the project root via go.mod).
	Stat(name string) (fs.FileInfo, error)
	// MkdirAll creates a directory and any missing parents.
	MkdirAll(path string, perm fs.FileMode) error
	// WriteFile writes data to name, creating or truncating it.
	WriteFile(name string, data []byte, perm fs.FileMode) error
	// ReadFile reads the named file.
	ReadFile(name string) ([]byte, error)
	// ReadDir lists directory entries.
	ReadDir(name string) ([]os.DirEntry, error)
	// RemoveAll removes a path and any children (used for scratch dirs).
	RemoveAll(path string) error
}

// OS is the production FS backed by the os package.
type OS struct{}

// Stat implements FS by calling os.Stat.
func (OS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }

// MkdirAll implements FS by calling os.MkdirAll.
func (OS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

// WriteFile implements FS by calling os.WriteFile.
func (OS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// ReadFile implements FS by calling os.ReadFile.
func (OS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

// ReadDir implements FS by calling os.ReadDir.
func (OS) ReadDir(name string) ([]os.DirEntry, error) { return os.ReadDir(name) }

// RemoveAll implements FS by calling os.RemoveAll.
func (OS) RemoveAll(path string) error { return os.RemoveAll(path) }
