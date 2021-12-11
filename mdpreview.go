package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/vui-chee/mdpreview/service"
	m "github.com/vui-chee/mdpreview/service/middleware"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		exitAfterUsage("Please enter a file.")
	}

	var filepath string = flag.Args()[0]
	if !isMarkdownFile(filepath) {
		exitAfterUsage("File must be a markdown document.")
	}

	service.Watch(filepath)
	service.Start(m.Args{Filepath: filepath})
}

func exitAfterUsage(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	os.Exit(1)
}

// Returns false if path entered is not a
// valid markdown file.
func isMarkdownFile(filepath string) bool {
	info, err := os.Lstat(filepath)
	if err != nil || info.IsDir() {
		return false
	}

	ext := path.Ext(info.Name())
	if ext != ".md" {
		return false
	}

	return true
}
