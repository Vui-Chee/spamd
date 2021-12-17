package sys

import (
	"fmt"
	"os"
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
