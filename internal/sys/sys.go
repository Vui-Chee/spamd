package sys

import (
	"runtime"
)

type Delegate interface {
	Start() error
}

type Commands map[string](Delegate)

func Exec(syscmd Commands) error {
	return syscmd[runtime.GOOS].Start()
}
