package service

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path"
)

type Message struct {
	Content string `json:"content"`
}

// To insert variables into `index.html`.
type InsertHTML struct {
	Filename string
}

// Need multiple channels for each connection, otherwise
// only a single connection will be notified of any changes.
var messageChannels = make(map[chan string]bool)

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
	githubMarkdownCSS, err := getEmbeddedBytes("build/styles.css")
	if err != nil {
		exitOnError(err.Error())
	}
	w.Header().Set("Content-Type", "text/css")
	w.Write(githubMarkdownCSS)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	mainHTML, err := getEmbeddedBytes("build/index.html")
	if err != nil {
		exitOnError(err.Error())
	}
	t := template.New("Main HTML template")
	t, _ = t.Parse(string(mainHTML))
	filepath := r.Context().Value("filepath").(string)
	data := InsertHTML{Filename: path.Base(filepath)}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func refreshContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	filepath := r.Context().Value("filepath").(string)

	// Create a new channel for each connection.
	singleChannel := make(chan string)
	messageChannels[singleChannel] = true

	for {
		select {
		case <-singleChannel:
			content, err := convertMarkdownToHTML(filepath)
			if err != nil {
				log.Fatalln(err)
				continue
			}
			w.Write(eventStreamFormat(string(content)))
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			delete(messageChannels, singleChannel)
			log.Println("User closed tab. This connection is closed.")
			return
		}
	}
}
