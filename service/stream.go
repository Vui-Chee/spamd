package service

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vui-chee/mdpreview/internal/sys"
	conf "github.com/vui-chee/mdpreview/service/config"
)

const (
	ENDLESS_LOOP = -1
)

type fileInfo struct {
	Lastmodifed time.Time
	Count       int
}

type FileWatcher struct {
	// Maps a relative filepath to the number browser tabs
	// that open it as well as the files last modified time.
	trackFiles map[string]*fileInfo
	// Need multiple channels for each connection, otherwise
	// only a single connection will be notified of any changes.
	messageChannels map[chan string]string

	// Use single crude lock over all shared data structures. For this
	// simple use case where only one user read/writes to markdown file,
	// it is not required to have a performant locking mechanism.
	lock sync.Mutex

	// Used in testing to control number of iterations of main listener loop.
	loops int
}

func (f *FileWatcher) RefreshContent(w http.ResponseWriter, r *http.Request) {
	// Get the path relative to the directory where the tool is run.
	// '+1' to skip the leading '/'.
	uri := r.URL.Path
	filepath := uri[len(conf.RefreshPrefix)+1:]

	// Create a new channel for each connection.
	singleChannel := make(chan string)

	func() {
		f.lock.Lock()
		defer f.lock.Unlock()

		// Pass the URI as a value to allow the watcher
		// to filter channels by file that has been modified.
		f.messageChannels[singleChannel] = filepath

		if _, ok := f.trackFiles[filepath]; ok {
			f.trackFiles[filepath].Count++
		} else {
			f.trackFiles[filepath] = &fileInfo{
				Count:       1,
				Lastmodifed: time.Time{}, // trigger the first page
			}
		}
	}()

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	for f.loops == ENDLESS_LOOP || f.loops > 0 {
		if f.loops > 0 {
			f.loops--
		}

		select {
		case <-singleChannel:
			if err := readAndSendMarkdown(w, filepath); err != nil {
				log.Fatalln(err)
				continue
			}
		case <-r.Context().Done():
			func() {
				f.lock.Lock()
				defer f.lock.Unlock()
				// Only decrement if key exists.
				if _, ok := f.trackFiles[filepath]; ok {
					f.trackFiles[filepath].Count--
					if f.trackFiles[filepath].Count <= 0 {
						delete(f.trackFiles, filepath)
					}
				}

				delete(f.messageChannels, singleChannel)
			}()

			log.Println("User closed tab. This connection is closed.")
			return
		}
	}
}

func (f *FileWatcher) Watch() {
	go func() {
		for {
			time.Sleep(300 * time.Millisecond) // 0.3s

			func() {
				f.lock.Lock()
				defer f.lock.Unlock()

				for filepath, info := range f.trackFiles {
					newModtime, err := sys.Modtime(filepath)
					if err != nil {
						log.Fatal(err)
						continue
					}

					if info.Lastmodifed != newModtime {
						fmt.Println("File was modified at: ", time.Now().Local())

						info.Lastmodifed = newModtime // update modified time

						for messageChannel, channelPath := range f.messageChannels {
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

func NewFileWatcher() *FileWatcher {
	return &FileWatcher{
		trackFiles:      make(map[string]*fileInfo),
		messageChannels: make(map[chan string]string),
		lock:            sync.Mutex{},

		loops: ENDLESS_LOOP,
	}
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
		if len(line) == 0 {
			// This is just a single newline.
			eventPayload = eventPayload + "data:\n"
		} else {
			eventPayload = eventPayload + "data:" + line + "\n"
		}
	}

	return []byte(eventPayload + "\n")
}
