package service

import (
	"fmt"
	"log"
	"net"
	"net/http"

	m "github.com/vui-chee/mdpreview/service/middleware"
)

func Start(args m.Args) {
	port, err := getFreePort()
	if err != nil {
		exitOnError("Failed to get port.")
	}

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		exitOnError(fmt.Sprintf("Failed to start server at %d.\n", port))
	}
	fmt.Printf("Server started at port %d.\n", port)

	openbrowser(fmt.Sprintf("http://localhost:%d", port))

	mux := http.NewServeMux()
	mux.HandleFunc("/content", currentPage)
	mux.HandleFunc("/styles", serveCSS)
	mux.HandleFunc("/", serveHTML)

	// Setup middlewares
	wrapper := m.NewArgsInjector(m.NewLogger(mux), args)

	// Start the blocking server loop.
	log.Fatal(http.Serve(l, wrapper))
}

// TODO: create file watching service
func Watch(filename string) {
	fmt.Println("Watching: ", filename)
}
