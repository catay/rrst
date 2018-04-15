package file

import (
	"os"
)

func IsRegularFile(name string) bool {
	fi, err := os.Stat(name)

	if err != nil {
		return false
	}

	if fi.Mode().IsRegular() {
		return true
	}
	return false
}

func IsDirectory(name string) bool {
	fi, err := os.Stat(name)

	if err != nil {
		return false
	}

	if fi.IsDir() {
		return true
	}
	return false
}
