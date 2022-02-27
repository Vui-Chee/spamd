package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/vui-chee/spamd/internal/browser"
	"github.com/vui-chee/spamd/internal/options"
	"github.com/vui-chee/spamd/internal/sys"
	"github.com/vui-chee/spamd/service"
)

const (
	version  = "0.1.1"
	protocol = "http://"
)

func main() {
	closeOnCtrlC()

	opts := options.ParseOptions()

	if opts.ShowVersion {
		fmt.Println(version)
		return
	}

	service.OverrideConfig(opts.Theme, opts.CodeStyle)

	l, err := service.Listen(opts.Port)
	if err != nil {
		sys.ErrorAndExit(err.Error())
	}
	baseUrl := protocol + l.Addr().String()

	browser.MassOpen(baseUrl, opts.NoBrowser)

	printAdditionalInfo(baseUrl)
	service.Start(l)
}

func closeOnCtrlC() {
	// Capture ctrl-c to close server.
	var interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Close all websocket connections before exiting.
	go func() {
		<-interrupt
		service.Shutdown()
		os.Exit(1)
	}()
}

func printAdditionalInfo(address string) {
	fmt.Printf(`Visit your markdown at %s/{path-to-markdown}.

{path-to-markdown} can be a relative path from current directory.
`, address)
}

func exitAfterUsage(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	os.Exit(1)
}
