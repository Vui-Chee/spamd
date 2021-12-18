package testing

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// This mocks the embedded fs used in the static handlers.
//go:embed mockfs
var MockFS embed.FS

var tempdir string

// Returns true if the last node in the subpath is a directory.
func pathIsDir(subpath string) bool {
	return subpath[len(subpath)-1] == os.PathSeparator && len(path.Base(subpath)) > 0
}

// Put all files/folders inside main folder.
func SetupFS(prefix string, subpaths []string) (string, error) {
	// Creates a temporary directory
	testdir, err := ioutil.TempDir(".", prefix)
	if err != nil {
		return "", fmt.Errorf("Failed to create test directory.")
	}

	for _, subpath := range subpaths {
		// Join subpath with folder containing test files.
		p := fmt.Sprintf("%s/%s", testdir, subpath)

		if pathIsDir(p) {
			// Create directory
			err = os.Mkdir(p, 0777)
			if err != nil {
				return testdir, fmt.Errorf("Failed to create dir: %s.", p)
			}
		} else {
			// Create file
			_, err := os.Create(p)
			if err != nil {
				return testdir, fmt.Errorf("Failed to create file %s.", p)
			}
		}
	}

	return testdir, nil
}

func Teardown(paths ...string) {
	for _, path := range paths {
		os.RemoveAll(path)
	}
}
