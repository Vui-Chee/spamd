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

func initRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/refresh", refreshContent)
	mux.HandleFunc("/styles", serveCSS)
	mux.HandleFunc("/", serveHTML)
}

func Listen() net.Listener {
	port, err := common.NextPort()
	if err != nil {
		sys.ErrorAndExit("Failed to get port.")
	}

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		sys.ErrorAndExit(fmt.Sprintf("Failed to start server at %d.\n", port))
	}
	fmt.Printf("Server started at port %d.\n", port)

	return l
}

func additionalCheck(path string) bool {
	if path == "/styles" {
		return true
	}

	var uri string
	refreshRegex, _ := regexp.Compile("^/refresh/.+")
	htmlRegex, _ := regexp.Compile("^/.+")
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

func Start(l net.Listener, args m.Args) {
	mux := m.RegexpHandler{
		AdditionalCheck: additionalCheck,
	}
	mux.HandleFunc(StylesPattern, serveCSS)
	mux.HandleFunc(RefreshPattern, refreshContent)
	mux.HandleFunc(AllElse, serveHTML)
	wrapper := m.NewLogger(&mux)

	log.Fatal(http.Serve(l, wrapper))
}

func Watch() {
	watchFile()
}
