package service

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
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

func getEmbeddedBytes(filepath string) ([]byte, error) {
	data, err := f.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func currentPage(w http.ResponseWriter, r *http.Request) {
	filepath := r.Context().Value("filepath").(string)

	content, err := convertMarkdownToHTML(filepath)
	if err != nil {
		log.Fatalln(err)
		return
	}

	// Send markdown contents in Json format.
	w.Header().Set("Content-Type", "application/json")

	type message struct {
		Content string `json:"content"`
	}
	msg := message{Content: string(content)}
	json.NewEncoder(w).Encode(msg)
}

func serveCSS(w http.ResponseWriter, r *http.Request) {
	githubMarkdownCSS, err := getEmbeddedBytes(fsPrefix + "/" + "styles.css")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Failed to read from styles.css"))
		return
	}
	w.Header().Set("Content-Type", "text/css")
	w.Write(githubMarkdownCSS)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	mainHTML, err := getEmbeddedBytes(fsPrefix + "/" + "index.html")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Failed to read from index.html"))
		return
	}

	t := template.New("Main HTML template")
	t, _ = t.Parse(string(mainHTML))
	filepath := r.Context().Value("filepath").(string)

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, map[string]string{"Filename": path.Base(filepath)})
}
