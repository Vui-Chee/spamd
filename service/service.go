package service

import (
	"fmt"
	"log"
	"net"
	"net/http"

	m "github.com/vui-chee/mdpreview/service/middleware"
)

func initRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/refresh", refreshContent)
	mux.HandleFunc("/content", currentPage)
	mux.HandleFunc("/styles", serveCSS)
	mux.HandleFunc("/", serveHTML)
}

func Listen() net.Listener {
	port, err := getFreePort()
	if err != nil {
		exitOnError("Failed to get port.")
	}

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		exitOnError(fmt.Sprintf("Failed to start server at %d.\n", port))
	}
	fmt.Printf("Server started at port %d.\n", port)

	return l
}

func Start(l net.Listener, args m.Args) {
	// Initialize routes
	mux := http.NewServeMux()
	initRoutes(mux)
	// Setup middlewares
	wrapper := m.NewArgsInjector(m.NewLogger(mux), args)
	// Start the blocking server loop.
	log.Fatal(http.Serve(l, wrapper))
}

func Watch(filepath string) {
	watchFile(filepath)
}
