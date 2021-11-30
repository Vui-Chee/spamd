package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/yuin/goldmark"
)

const (
	PORT = 3001
)

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func startServer(filebytes bytes.Buffer) {
	fmt.Printf("Starting server at port %d\n", PORT)
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", PORT))
	if err != nil {
		log.Fatal(err)
	}

	openbrowser(fmt.Sprintf("http://localhost:%d", PORT))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, filebytes.String())
	})
	// Start the blocking server loop.
	log.Fatal(http.Serve(l, mux))
}

func main() {
	if len(os.Args) != 2 {
		panic("Please enter a markdown file.")
	}

	var filename string = os.Args[1]

	// Read the markdown file
	dat, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalln("Error reading file.")
		os.Exit(1)
	}

	// Parse the file data (in bytes) into a buffer.
	var filebytes bytes.Buffer
	if err := goldmark.Convert(dat, &filebytes); err != nil {
		panic(err)
	}

	startServer(filebytes)
}
