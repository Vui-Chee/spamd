package service

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/vui-chee/mdpreview/internal/sys"
)

// Need multiple channels for each connection, otherwise
// only a single connection will be notified of any changes.
var messageChannels = make(map[chan string]bool)

func refreshContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context().Value("filepath")
	if ctx == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("refreshContent: ctx filepath is <nil>."))
		return
	}

	filepath := ctx.(string)

	// Create a new channel for each connection.
	singleChannel := make(chan string)
	messageChannels[singleChannel] = true

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Qn. How does this know which channel to listen without the filepath???
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
		modtime, err := sys.Modtime(filepath)
		if err != nil {
			log.Fatal(err)
			return
		}

		for {
			time.Sleep(300 * time.Millisecond) // 0.3s

			newModtime, err := sys.Modtime(filepath)
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

// In order for the client side to receive server triggered
// event messages, the data sent must be formatted in a specific
// way, otherwise, the data will be dropped. For event streams,
// messages within this stream are represented as a sequence
// of bytes separated by a newline. The data must also be encoded
// in UTF-8.
//
// See https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
// for more information.
func eventStreamFormat(data string) []byte {
	if len(data) <= 0 {
		return []byte("")
	}

	var eventPayload string
	dataLines := strings.Split(data, "\n")
	for _, line := range dataLines {
		if len(line) > 0 {
			eventPayload = eventPayload + "data: " + line + "\n"
		}
	}

	if len(eventPayload) <= 0 {
		return []byte("")
	}

	return []byte(eventPayload + "\n")
}
