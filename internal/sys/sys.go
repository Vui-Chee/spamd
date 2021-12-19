package sys

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"
)

type Delegate interface {
	Start() error
}

type Commands map[string](Delegate)

func Exec(syscmd Commands) error {
	return syscmd[runtime.GOOS].Start()
}

func ErrorAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	os.Exit(1)
}

// Returns the last modified time on file.
func Modtime(filename string) (time.Time, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return time.Time{}, err
	}

	return info.ModTime(), nil
}

func Exists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

// Returns false if path entered is not a
// valid markdown file.
func IsFileWithExt(filepath string, targetExt string) bool {
	info, err := os.Lstat(filepath)
	if err != nil || info.IsDir() {
		return false
	}

	ext := path.Ext(info.Name())
	if ext != targetExt {
		return false
	}

	return true
}
