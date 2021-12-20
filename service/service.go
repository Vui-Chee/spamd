package service

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"

	"github.com/vui-chee/mdpreview/internal/common"
	"github.com/vui-chee/mdpreview/internal/sys"
	m "github.com/vui-chee/mdpreview/service/middleware"
)

const (
	RefreshPattern = "^/refresh/.+"
	StylesPattern  = "/styles"
	AllElse        = "^/.+"
)

func Listen() net.Listener {
	port, err := common.NextPort()
	if err != nil {
		sys.ErrorAndExit("Failed to get port.")
	}

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		sys.ErrorAndExit(fmt.Sprintf("Failed to start server at %d.\n", port))
	}

	return l
}

func redirectIfNotMarkdown(path string) bool {
	if path == StylesPattern {
		return true
	}

	var uri string
	refreshRegex, _ := regexp.Compile(RefreshPattern)
	htmlRegex, _ := regexp.Compile(AllElse)
	if refreshRegex.MatchString(path) {
		uri = path[len("/refresh"):]
	} else if htmlRegex.MatchString(path) {
		uri = path
	}

	cwd, _ := os.Getwd()
	if !sys.IsFileWithExt(cwd+uri, ".md") {
		return false
	}

	return true
}

func Start(l net.Listener) {
	watcher := NewFileWatcher()
	mux := m.RegexpHandler{
		AdditionalCheck: redirectIfNotMarkdown,
	}
	mux.HandleFunc(StylesPattern, serveCSS)
	mux.HandleFunc(RefreshPattern, watcher.RefreshContent)
	mux.HandleFunc(AllElse, serveHTML)
	wrapper := m.NewLogger(&mux)

	// Must call this before main thread is blocked
	// http.Serve.
	watcher.Watch()

	log.Fatal(http.Serve(l, wrapper))
}
