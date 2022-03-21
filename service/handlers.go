package service

import (
	"embed"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	conf "github.com/vui-chee/spamd/service/config"
)

var (
	//go:embed frontend
	f embed.FS

	// Folder where static files are stored (relative to this directory).
	// This variable is overwritten during testing with the folder
	// where the static files are actually stored.
	fsPrefix string = "frontend"
)

func serveLocalImage(w http.ResponseWriter, r *http.Request) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory. %+v\n", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Opens the image file relative to current directory.
	img, err := os.Open(path.Join(wd, r.URL.Path))
	if err != nil {
		log.Printf("%+v\n", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer img.Close()

	contentType := "image/"

	ext := path.Ext(r.URL.Path)
	if len(ext) > 0 && ext[0] == '.' {
		// skip '.'
		ext = ext[1:]
	}

	// svg+xml
	if ext == "svg" {
		contentType += ext + "+xml"
	}

	w.Header().Set("Content-Type", contentType)
	io.Copy(w, img)
}

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
	t.Execute(w, map[string]string{"Filename": path.Base(r.URL.Path),
		"URI":           r.URL.Path,
		"Theme":         serviceConfig.Theme,
		"RefreshPrefix": conf.RefreshPrefix,
		"StylesPrefix":  conf.StylesPrefix,
	})
}
