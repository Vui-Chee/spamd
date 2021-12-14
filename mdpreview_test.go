package main

import (
	"os"
	"testing"
)

func TestValidMarkdownFile(t *testing.T) {
	cwd, _ := os.Getwd()

	cases := []string{
		"README.md",                       // just filename
		"examples/large-readme.md",        // relative path
		cwd + "/examples/large-readme.md", // absolute path to file
	}

	for _, filepath := range cases {
		got := isMarkdownFile(filepath)
		if !got {
			t.Errorf("isMarkdownFile(\"%s\") should return true.", filepath)
		}
	}

}

func TestInvalidMarkdownFile(t *testing.T) {
	cwd, _ := os.Getwd()

	cases := []string{
		"examples", // directory
		".",        // relative path(s)
		"../",
		"foobar", // non-existent node
		cwd,      // full path (points to root directory)
		"go.mod", // invalid extension
	}

	for _, filepath := range cases {
		got := isMarkdownFile(filepath)
		if got {
			t.Errorf("isMarkdownFile(\"%s\") should return false.", filepath)
		}
	}
}
