package service

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Need multiple channels for each connection, otherwise
// only a single connection will be notified of any changes.
var messageChannels = make(map[chan string]bool)

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

func watchFile(filepath string) {
	go func() {
		modtime, err := getFileModtime(filepath)
		if err != nil {
			log.Fatal(err)
			return
		}

		for {
			time.Sleep(300 * time.Millisecond) // 0.3s

			newModtime, err := getFileModtime(filepath)
			if err != nil {
				log.Fatal(err)
				continue
			}

			if modtime != newModtime {
				fmt.Println("File was modified at: ", time.Now().Local())
				modtime = newModtime
				for messageChannel := range messageChannels {
					messageChannel <- newModtime.String()
				}
			}
		}
	}()
}
