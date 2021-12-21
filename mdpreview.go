package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/vui-chee/mdpreview/internal/browser"
	"github.com/vui-chee/mdpreview/internal/sys"
	"github.com/vui-chee/mdpreview/service"
)

const (
	defaultMarkdown = "README.md"

	protocol = "http://"

	usage = `Usage: mdpreview <path-to-markdown>
`
)

// When applied, these value(s) will override as existing configuration.
var (
	p = flag.Int("p", -1, "port")
)

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	var filepath string = defaultMarkdown
	var l net.Listener = service.Listen(*p)
	defer l.Close()

	if flag.NArg() >= 1 {
		filepath = flag.Args()[0]

		if !sys.IsFileWithExt(filepath, ".md") {
			exitAfterUsage("File must be a markdown document.")
		}
	}

	if (flag.NArg() >= 1 && sys.IsFileWithExt(filepath, ".md")) || sys.Exists(filepath) {
		sys.Exec(browser.Commands(protocol + l.Addr().String() + "/" + filepath))
	}

	fmt.Printf("Visit your markdown at %s/{path-to-markdown}.\n\n", protocol+l.Addr().String())
	fmt.Println("{path-to-markdown} can be a relative path from current directory.")

	service.Start(l)
}

func exitAfterUsage(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	os.Exit(1)
}
