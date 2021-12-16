package sys

import (
	"runtime"
	"testing"
)

type TestDelegate struct {
	count int
}

func (t *TestDelegate) Start() error {
	t.count++

	return nil
}

func stubCommands() Commands {
	return Commands{
		"linux":   &TestDelegate{},
		"windows": &TestDelegate{},
		"darwin":  &TestDelegate{},
	}
}

func TestExecuteByOS(t *testing.T) {
	commands := stubCommands()
	Exec(commands)
	count := commands[runtime.GOOS].(*TestDelegate).count

	if count != 1 {
		t.Errorf("Exec() should be ran once. Got %d", count)
	}
}
