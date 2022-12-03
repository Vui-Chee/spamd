package sys

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	testtools "spamd/internal/testing"
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

var testdir string

const (
	prefix = "spamd-test-"
)

var testFiles = []string{
	"TEST_README.md",
	"go.mod",
}

var testSubdir = "example"

func init() {
	var err error

	_, err = os.Create(testFiles[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create top-level test README.md.")
		os.Exit(1)
	}

	subpaths := []string{
		"/TEST_README.md",
		"/go.mod",
		"/example/",
	}

	testdir, err = testtools.SetupFS(prefix, subpaths)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		testtools.Teardown(testdir)
		os.Exit(1)
	}
}

func TestValidMarkdownFile(t *testing.T) {
	defer testtools.Teardown(testdir, testFiles[0])

	cwd, _ := os.Getwd()
	cases := []string{
		testFiles[0],                 // just filename
		testdir + "/" + testFiles[0], // relative path
		cwd + "/" + testFiles[0],     // absolute path to file
	}

	for _, filepath := range cases {
		got := IsFileWithExt(filepath, ".md")
		if !got {
			t.Errorf("isMarkdownFile(\"%s\") should return true.", filepath)
		}
	}
}

func TestInvalidMarkdownFile(t *testing.T) {
	defer testtools.Teardown(testdir, testFiles[0])

	cwd, _ := os.Getwd()
	cases := []string{
		".", // relative path(s)
		"../",
		"foobar.md",                  // non-existent file
		cwd,                          // full path (points to root directory)
		testdir + "/" + testSubdir,   // directory
		testdir + "/" + testFiles[1], // invalid extension
	}

	for _, filepath := range cases {
		got := IsFileWithExt(filepath, ".md")
		if got {
			t.Errorf("isMarkdownFile(\"%s\") should return false.", filepath)
		}
	}
}
