package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vui-chee/mdpreview/internal/browser"
	"github.com/vui-chee/mdpreview/internal/sys"
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
	if !sys.IsFileWithExt(filepath, ".md") {
		exitAfterUsage("File must be a markdown document.")
	}

	l := service.Listen()

	// Open address in browser based on system.
	sys.Exec(browser.Commands(protocol + l.Addr().String() + "/" + filepath))

	service.Watch()
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
