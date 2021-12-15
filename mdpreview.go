package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/vui-chee/mdpreview/internal/browser"
	"github.com/vui-chee/mdpreview/service"
	m "github.com/vui-chee/mdpreview/service/middleware"
)

const (
	protocol = "http://"
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

	l := service.Listen()

	// Open address in browser based on system.
	browser.Open(protocol + l.Addr().String())

	service.Watch(filepath)
	service.Start(l, m.Args{Filepath: filepath})
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
