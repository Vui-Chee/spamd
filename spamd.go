package main

import (
	"os"
	"os/signal"

	"github.com/vui-chee/spamd/internal/options"
	"github.com/vui-chee/spamd/service"
)

const (
	version = "0.1.1"
)

func main() {
	closeOnCtrlC()
	service.Run(options.ParseOptions(), version)
}

func closeOnCtrlC() {
	var interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		// Close all websocket connections before exiting.
		service.Shutdown()
		os.Exit(1)
	}()
}
