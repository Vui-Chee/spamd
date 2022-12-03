package service

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"

	"spamd/internal/browser"
	"spamd/internal/options"
	"spamd/internal/sys"
	"spamd/service/config"
	"spamd/service/middleware"
)

const (
	TOOL_NAME = "spamd"

	// Remaining unmatched routes go to default html handler.
	// For other static routes, see config package.
	AllElse = "^/.+"

	// Everything is served locally.
	protocol = "http://"
)

// Set the configs for this service as a global,
// but only accessible within the service package.
var (
	serviceConfig *config.ServiceConfig

	watcher *FileWatcher
)

func init() {
	var err error
	serviceConfig, err = config.ReadConfigFromFile("." + TOOL_NAME)
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
	mux := middleware.RegexpHandler{
		AdditionalCheck: redirectIfNotMarkdown,
	}
	mux.HandleFunc(config.StylesPrefix, serveCSS)
	mux.HandleFunc(config.ImageRegex, serveLocalImage)
	mux.HandleFunc(config.RefreshPattern(), watcher.RefreshContent)
	mux.HandleFunc(AllElse, serveHTML)
	wrapper := middleware.NewLogger(&mux)

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
	if path == config.StylesPrefix {
		return true
	}

	var uri string
	refreshRegex, _ := regexp.Compile(config.RefreshPattern())
	htmlRegex, _ := regexp.Compile(AllElse)
	if refreshRegex.MatchString(path) {
		uri = path[len(config.RefreshPrefix):]
	} else if htmlRegex.MatchString(path) {
		uri = path
	}

	imageRegex, _ := regexp.Compile(config.ImageRegex)
	cwd, _ := os.Getwd()
	// Skip non-markdown and non-image files.
	if !sys.IsFileWithExt(cwd+uri, ".md") && !imageRegex.Match([]byte(uri)) {
		return false
	}

	return true
}

func printAdditionalInfo(address string) {
	fmt.Printf(`Visit your markdown at %s/{path-to-markdown}.

{path-to-markdown} can be a relative path from current directory.
`, address)
}

func Run(opts *options.Options, version string) {
	if opts.ShowVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	OverrideConfig(opts.Theme, opts.CodeStyle)

	l, err := Listen(opts.Port)
	if err != nil {
		sys.ErrorAndExit(err.Error())
	}
	baseUrl := protocol + l.Addr().String()

	browser.MassOpen(baseUrl, opts.NoBrowser)

	printAdditionalInfo(baseUrl)
	Start(l)
}
