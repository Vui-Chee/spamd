package service

import (
	"encoding/json"
	"log"
	"net/http"
)

type Message struct {
	Content string `json:"content"`
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
	msg := Message{Content: string(content)}
	json.NewEncoder(w).Encode(msg)
}

func serveCSS(w http.ResponseWriter, r *http.Request) {
	githubMarkdownCSS, err := getEmbeddedBytes("frontend/node_modules/github-markdown-css/github-markdown.css")
	if err != nil {
		exitOnError(err.Error())
	}
	w.Header().Set("Content-Type", "text/css")
	w.Write(githubMarkdownCSS)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	mainHTML, err := getEmbeddedBytes("frontend/index.html")
	if err != nil {
		exitOnError(err.Error())
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(mainHTML)
}
