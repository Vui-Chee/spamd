package service

import (
	"embed"
	"html/template"
	"net/http"
	"path"
)

var (
	//go:embed build
	f embed.FS

	// Folder where static files are stored (relative to this directory).
	// This variable is overwritten during testing with the folder
	// where the static files are actually stored.
	fsPrefix string = "build"
)

func serveCSS(w http.ResponseWriter, r *http.Request) {
	githubMarkdownCSS, err := f.ReadFile(fsPrefix + "/" + "styles.css")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Failed to read from styles.css"))
		return
	}
	w.Header().Set("Content-Type", "text/css")
	w.Write(githubMarkdownCSS)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	mainHTML, err := f.ReadFile(fsPrefix + "/" + "index.html")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Failed to read from index.html"))
		return
	}

	t := template.New("Main HTML template")
	t, _ = t.Parse(string(mainHTML))

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, map[string]string{"Filename": path.Base(r.URL.Path), "URI": r.URL.Path, "Theme": serviceConfig.Theme})
}
