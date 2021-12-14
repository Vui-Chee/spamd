package main

import (
	"fmt"
	"os"
	"testing"
)

const (
	testdir = "tests"
)

var testFiles = []string{
	"TEST_README.md",
	"go.mod",
}

var testSubdir = "example"

func setupFiles() (bool, string) {
	err := os.Mkdir(testdir, 0777)
	if err != nil {
		return false, "Failed to create test directory."
	}

	// Create top-level test readme file.
	_, err = os.Create(testFiles[0])
	if err != nil {
		return false, fmt.Sprintf("Failed to create file %s.", testFiles[0])
	}

	// Files created within `testdir`.
	for _, file := range testFiles {
		path := fmt.Sprintf("%s/%s", testdir, file)
		_, err := os.Create(path)
		if err != nil {
			return false, fmt.Sprintf("Failed to create file %s.", path)
		}
	}

	// Create test sub-directory.
	path := fmt.Sprintf("%s/%s", testdir, testSubdir)
	err = os.Mkdir(path, 0777)
	if err != nil {
		return false, fmt.Sprintf("Failed to create dir: %s.", path)
	}

	return true, ""
}

func setup() {
	got, msg := setupFiles()
	if !got {
		fmt.Fprintf(os.Stderr, msg)
		teardown()
		os.Exit(1) // stop running test if setup failed
	}
}

func teardown() {
	os.Remove(testFiles[0])
	os.RemoveAll(testdir)
}

func TestValidMarkdownFile(t *testing.T) {
	setup()
	defer teardown()

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
}

func TestInvalidMarkdownFile(t *testing.T) {
	setup()
	defer teardown()

	cwd, _ := os.Getwd()
	cases := []string{
		testSubdir, // directory
		".",        // relative path(s)
		"../",
		"foobar.md",                  // non-existent file
		cwd,                          // full path (points to root directory)
		testdir + "/" + testFiles[1], // invalid extension
	}

	for _, filepath := range cases {
		got := isMarkdownFile(filepath)
		if got {
			t.Errorf("isMarkdownFile(\"%s\") should return false.", filepath)
		}
	}
}
