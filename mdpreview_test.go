package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

const (
	prefix = "mdpreview-test-"
)

var testFiles = []string{
	"TEST_README.md",
	"go.mod",
}

var testSubdir = "example"

func setupFiles() (bool, string, string) {
	testdir, err := ioutil.TempDir(".", prefix)
	if err != nil {
		return false, "Failed to create test directory.", ""
	}

	// Create top-level test readme file.
	_, err = os.Create(testFiles[0])
	if err != nil {
		return false, fmt.Sprintf("Failed to create file %s.", testFiles[0]), testdir
	}

	// Files created within `testdir`.
	for _, file := range testFiles {
		path := fmt.Sprintf("%s/%s", testdir, file)
		_, err := os.Create(path)
		if err != nil {
			return false, fmt.Sprintf("Failed to create file %s.", path), testdir
		}
	}

	// Create test sub-directory.
	path := fmt.Sprintf("%s/%s", testdir, testSubdir)
	err = os.Mkdir(path, 0777)
	if err != nil {
		return false, fmt.Sprintf("Failed to create dir: %s.", path), testdir
	}

	return true, "", testdir
}

func setup() (string, string) {
	got, msg, testdir := setupFiles()
	if !got {
		return testdir, msg
	}

	return testdir, ""
}

func teardown(testdir string) {
	os.Remove(testFiles[0])
	os.RemoveAll(testdir)
}

func setupAndTeardown(run func(string)) {
	testdir, errMsg := setup()
	if len(errMsg) > 0 {
		fmt.Fprintf(os.Stderr, errMsg)
		teardown(testdir)
		os.Exit(1)
	}
	defer teardown(testdir)
	run(testdir)
}

func TestValidMarkdownFile(t *testing.T) {
	setupAndTeardown(func(testdir string) {
		cwd, _ := os.Getwd()
		cases := []string{
			testFiles[0],                 // just filename
			testdir + "/" + testFiles[0], // relative path
			cwd + "/" + testFiles[0],     // absolute path to file
		}

		for _, filepath := range cases {
			got := isMarkdownFile(filepath)
			if !got {
				t.Errorf("isMarkdownFile(\"%s\") should return true.", filepath)
			}
		}
	})
}

func TestInvalidMarkdownFile(t *testing.T) {
	setupAndTeardown(func(testdir string) {
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
			got := isMarkdownFile(filepath)
			if got {
				t.Errorf("isMarkdownFile(\"%s\") should return false.", filepath)
			}
		}
	})
}
