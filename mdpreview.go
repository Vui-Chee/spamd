package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vui-chee/mdpreview/service"
	m "github.com/vui-chee/mdpreview/service/middleware"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		exitAfterUsage("Please enter a file.")
	}

	var filepath string = flag.Args()[0]

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
