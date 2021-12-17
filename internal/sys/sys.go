package sys

import (
	"fmt"
	"os"
	"runtime"
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
