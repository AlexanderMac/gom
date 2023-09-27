package gom

import (
	"io/fs"
	"os"
	"path/filepath"
)

type simpleFS interface {
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

type gomFS struct{}

func (gomFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(filepath.FromSlash(name))
}

func (gomFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.FromSlash(name))
}
