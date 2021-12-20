package service

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vui-chee/mdpreview/internal/sys"
)

type fileInfo struct {
	Lastmodifed time.Time
	Count       int
}

// Maps a relative filepath to the number browser tabs
// that open it as well as the files last modified time.
var fileTracker = make(map[string]*fileInfo)

// Allows many threads to read `fileTracker` and only 1 write
// at any one time.
var lock = sync.RWMutex{}

// Need multiple channels for each connection, otherwise
// only a single connection will be notified of any changes.
var messageChannels = make(map[chan string]string)

func refreshContent(w http.ResponseWriter, r *http.Request) {
	// Get the path relative to the directory where the tool is run.
	// '+1' to skip the leading '/'.
	uri := r.URL.Path
	filepath := uri[len("/refresh")+1:]

	// Create a new channel for each connection.
	singleChannel := make(chan string)
	// Pass the URI as a value to allow the watcher
	// to filter channels by file that has been modified.
	messageChannels[singleChannel] = filepath

	// Create a critical block.
	func() {
		lock.RLock()
		defer lock.RUnlock()

		if _, ok := fileTracker[filepath]; ok { // read
			fileTracker[filepath].Count++ // write
		} else {
			fileTracker[filepath] = &fileInfo{ // write
				Count:       1,
				Lastmodifed: time.Now(),
			}
		}
	}()

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// First time create channel, also sent first page.
	if err := readAndSendMarkdown(w, filepath); err != nil {
		log.Fatalln(err)
		return
	}

	for {
		select {
		case <-singleChannel:
			if err := readAndSendMarkdown(w, filepath); err != nil {
				log.Fatalln(err)
				continue
			}
		case <-r.Context().Done():
			func() {
				lock.RLock()
				defer lock.RUnlock()
				// Only decrement if key exists.
				if _, ok := fileTracker[filepath]; ok {
					fileTracker[filepath].Count--
					if fileTracker[filepath].Count <= 0 {
						delete(fileTracker, filepath)
					}
				}
			}()

			delete(messageChannels, singleChannel)

			log.Println("User closed tab. This connection is closed.")
			return
		}
	}
}

func watchFile() {
	go func() {
		for {
			time.Sleep(300 * time.Millisecond) // 0.3s

			func() {
				lock.RLock()
				defer lock.RUnlock()

				for filepath, info := range fileTracker { // read
					newModtime, err := sys.Modtime(filepath)
					if err != nil {
						log.Fatal(err)
						continue
					}

					if info.Lastmodifed != newModtime {
						fmt.Println("File was modified at: ", time.Now().Local())

						info.Lastmodifed = newModtime // update modified time (write)

						for messageChannel, channelPath := range messageChannels {
							// Only write to channels belonging to filepath.
							if filepath == channelPath {
								messageChannel <- newModtime.String()
							}
						}
					}
				}
			}()

		}
	}()
}

func readAndSendMarkdown(w http.ResponseWriter, filepath string) error {
	content, err := convertMarkdownToHTML(filepath)
	if err != nil {
		return err
	}
	w.Write(eventStreamFormat(string(content)))
	w.(http.Flusher).Flush()
	return nil
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
