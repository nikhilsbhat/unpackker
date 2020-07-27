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
