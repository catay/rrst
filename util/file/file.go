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

// FirstExistingFile returns the first existing filename in the list,
// if none found, return the last entry by default even if it not exists.
func FirstExistingFile(list []string) string {
	var saved string
	for i := range list {
		saved = list[i]
		_, err := os.Stat(list[i])
		if os.IsNotExist(err) {
			continue
		}
		break
	}
	return saved
}
