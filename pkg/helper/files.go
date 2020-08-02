// Package helper consists of supporting libraries for Unpackker.
package helper

import (
	"os"
	"path/filepath"
)

// SplitBasePath returns file and filepath split
func SplitBasePath(path string, pathType ...string) string {
	dir, filepath := filepath.Split(path)
	if len(pathType) == 0 {
		return filepath
	}
	return dir
}

// Statfile helps in stating the file( checking existence of file or folder)
func Statfile(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// CreateFile creates the file with execution permission and it is required to execute the client stub of Unpackker.
func CreateFile(name string) (*os.File, error) {
	if !Statfile(filepath.Dir(name)) {
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// OpenFile opens the file so that the data of it can ne read.
func OpenFile(name string) (*os.File, error) {
	file, err := os.Open("notes.txt")
	if err != nil {
		return nil, err
	}
	return file, nil
}
