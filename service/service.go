package service

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"

	"github.com/vui-chee/mdpreview/internal/sys"
	conf "github.com/vui-chee/mdpreview/service/config"
	m "github.com/vui-chee/mdpreview/service/middleware"
)

const (
	TOOL_NAME = "mdpreview"

	// Remaining unmatched routes go to default html handler.
	// For other static routes, see config package.
	AllElse = "^/.+"
)

// Set the configs for this service as a global,
// but only accessible within the service package.
var (
	serviceConfig *conf.ServiceConfig

	watcher *FileWatcher
)

func init() {
	var err error
	serviceConfig, err = conf.ReadConfigFromFile("." + TOOL_NAME)
	if err != nil {
		sys.ErrorAndExit(err.Error())
	}
}

func OverrideConfig(theme string, codeBlockStyle string) {
	serviceConfig.SetTheme(theme)
	err := serviceConfig.SetCodeBlockTheme(codeBlockStyle)
	if err != nil {
		sys.ErrorAndExit(err.Error())
	}
}

func Listen(port int) (net.Listener, error) {
	var err error

	if port == 0 {
		port = serviceConfig.Port
	}

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to start server at %d.\n", port))
	}

	return l, nil
}

func Start(l net.Listener) {
	watcher = NewFileWatcher(false)
	mux := m.RegexpHandler{
		AdditionalCheck: redirectIfNotMarkdown,
	}
	mux.HandleFunc(conf.StylesPrefix, serveCSS)
	mux.HandleFunc(conf.ImageRegex, serveLocalImage)
	mux.HandleFunc(conf.RefreshPattern(), watcher.RefreshContent)
	mux.HandleFunc(AllElse, serveHTML)
	wrapper := m.NewLogger(&mux)

	// Must call this before main thread is blocked
	// http.Serve.
	watcher.harness.loops = ENDLESS_LOOP
	watcher.Watch()

	log.Fatal(http.Serve(l, wrapper))
}

func Shutdown() {
	fmt.Println("Shutting down server...")
	watcher.CloseAllConn()
	fmt.Println("Server has been shut down.")
}

func redirectIfNotMarkdown(path string) bool {
	if path == conf.StylesPrefix {
		return true
	}

	var uri string
	refreshRegex, _ := regexp.Compile(conf.RefreshPattern())
	htmlRegex, _ := regexp.Compile(AllElse)
	if refreshRegex.MatchString(path) {
		uri = path[len(conf.RefreshPrefix):]
	} else if htmlRegex.MatchString(path) {
		uri = path
	}

	imageRegex, _ := regexp.Compile(conf.ImageRegex)
	cwd, _ := os.Getwd()
	// Skip non-markdown and non-image files.
	if !sys.IsFileWithExt(cwd+uri, ".md") && !imageRegex.Match([]byte(uri)) {
		return false
	}

	return true
}
