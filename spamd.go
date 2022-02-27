package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/vui-chee/spamd/internal/browser"
	"github.com/vui-chee/spamd/internal/sys"
	"github.com/vui-chee/spamd/service"
)

const (
	version         = "0.1.1"
	defaultMarkdown = "README.md"
	protocol        = "http://"
	usage           = `Usage: spamd [options...] <path-to-markdown>

Options:
	-p Port number (fixed port, otherwise a RANDOM port is supplied)
	-t Display markdown HTML in "dark" or "light" theme. (default: light)
	-c The style you want to apply to your code blocks. (default: monokai)
	-nb Do not open browser if this is set true (default: false)
        -v Display version and exit

Additionally, if you want to persist any of this configs, you can
create a .spamd JSON file at your ROOT directory containing:

	{
	  "theme": "dark",
	  "codeblock": "fruity",
	  "port": 3000
	}

This is just an example. You can change/omit any of the fields.
`
)

// When applied, these value(s) will override as existing configuration.
var (
	port        = flag.Int("p", 0, "port")
	showVersion = flag.Bool("v", false, "Version")

	// These have default empty string values as ServiceConfig will supply
	// the defaults.
	theme     = flag.String("t", "", "Change light/dark theme.")
	codestyle = flag.String("c", "", "Change the code block style.")
	nobrowser = flag.Bool("nb", false, "Use this option to disable open browser on start.")
)

func main() {
	// Capture ctrl-c to close server.
	var interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	flag.Usage = func() {
		sys.Eprintf(usage)
	}

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	service.OverrideConfig(*theme, *codestyle)

	var filepath string = defaultMarkdown
	l, err := service.Listen(*port)
	if err != nil {
		sys.ErrorAndExit(err.Error())
	}

	if flag.NArg() >= 1 {
		for i := 0; i < len(flag.Args()); i++ {
			filepath := flag.Args()[i]

			if !sys.IsFileWithExt(filepath, ".md") {
				sys.Eprintf("%s is not a markdown document.\n", filepath)
			} else if !sys.Exists(filepath) {
				sys.Eprintf("%s does not exist.\n", filepath)
			} else {
				if !*nobrowser {
					go func() {
						sys.Exec(browser.Commands(protocol + l.Addr().String() + "/" + filepath))
					}()
				}
			}
		}
	} else {
		if !*nobrowser && sys.IsFileWithExt(filepath, ".md") && sys.Exists(filepath) {
			sys.Exec(browser.Commands(protocol + l.Addr().String() + "/" + filepath))
		}
	}

	fmt.Printf("Visit your markdown at %s/{path-to-markdown}.\n\n", protocol+l.Addr().String())
	fmt.Println("{path-to-markdown} can be a relative path from current directory.")

	// Close all websocket connections before exiting.
	go func() {
		<-interrupt
		service.Shutdown()
		os.Exit(1)
	}()

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
