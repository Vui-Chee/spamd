package service

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vui-chee/mdpreview/internal/sys"
	conf "github.com/vui-chee/mdpreview/service/config"
)

const (
	ENDLESS_LOOP = -1

	write_success = "success"
	error_read    = "error_read"
)

// This struct is used to store all information used during testing.
type testHarness struct {
	// Used in testing to control number of iterations of main listener loop.
	loops int
	// Currently used to wait for main loop (RefreshContent) to start before
	// writing to test file.
	wg *sync.WaitGroup
	// Determine whether to use wg or not.
	useWaitGroup bool
}

func NewTestHarness() testHarness {
	return testHarness{
		loops:        1,
		wg:           new(sync.WaitGroup),
		useWaitGroup: false,
	}
}

type conn struct {
	Ch chan string
}

func NewConn() *conn {
	return &conn{
		Ch: make(chan string),
	}
}

type connCluster struct {
	Lastmodifed time.Time
	conns       []*conn
}

func NewCluster(filepath string) *connCluster {
	modtime, err := sys.Modtime(filepath)
	if err != nil {
		return nil
	}

	return &connCluster{
		Lastmodifed: modtime,
	}
}

type FileWatcher struct {
	// Represents a set of connections per filepath (key).
	files map[string]*connCluster

	// Use single crude lock over all shared data structures. For this
	// simple use case where only one user read/writes to markdown file,
	// it is not required to have a performant locking mechanism.
	lock sync.Mutex

	// Number of milliseconds between each subsequent file reads.
	watchInv time.Duration

	// Initialized during testing.
	harness testHarness
}

func (f *FileWatcher) AddConn(filepath string, modtime time.Time) *conn {
	f.lock.Lock()
	defer f.lock.Unlock()

	newConn := NewConn()
	cluster, ok := f.files[filepath]
	if !ok {
		f.files[filepath] = &connCluster{
			Lastmodifed: modtime,
			conns: []*conn{
				newConn,
			},
		}
	} else {
		cluster.conns = append(cluster.conns, newConn)
	}

	return newConn
}

func (f *FileWatcher) DeleteConn(filepath string, targetConn *conn) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	var index int = -1
	cluster, ok := f.files[filepath]
	if !ok {
		// File does not exist in map.
		return fmt.Errorf("%s not found in map.", filepath)
	}

	for i, conn := range cluster.conns {
		if conn == targetConn {
			index = i
			break
		}
	}

	// Connection not found.
	if index == -1 {
		return fmt.Errorf("Connection %v not found for %s", targetConn, filepath)
	}

	updatedConnections := append(cluster.conns[:index], cluster.conns[index+1:]...)
	if len(updatedConnections) == 0 {
		// No more connections to this file, so drop key-value pair.
		delete(f.files, filepath)
	} else {
		cluster.conns = updatedConnections
	}

	return nil
}

func (f *FileWatcher) RefreshContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get the path relative to the directory where the tool is run.
	// '+1' to skip the leading '/'.
	uri := r.URL.Path
	filepath, err := filepath.EvalSymlinks(uri[len(conf.RefreshPrefix)+1:])
	if err != nil {
		w.Write([]byte("event:userdisconnect\n\n"))
		return
	}

	modtime, err := sys.Modtime(filepath)
	if err != nil {
		w.Write([]byte("event:userdisconnect\n\n"))
		return
	}
	conn := f.AddConn(filepath, modtime)

	// Read first page
	if err := readAndSendMarkdown(w, filepath); err != nil {
		log.Fatalln(err)
		return
	}

	for {
		// Used during testing only.
		if f.harness.loops > 0 {
			f.harness.loops--
		} else if f.harness.loops == 0 {
			break
		}

		select {
		case msg := <-conn.Ch:
			// During such error, file will be deleted from
			// trackFiles & messageChannels during Watch().
			if msg == error_read {
				w.Write([]byte("event:userdisconnect\n\n"))
				return
			}

			if msg == write_success {
				if err := readAndSendMarkdown(w, filepath); err != nil {
					log.Fatalln(err)
					continue
				}
			}
		case <-r.Context().Done():
			f.DeleteConn(filepath, conn)
			log.Println("User closed tab. This connection is closed.")
			return
		}
	}
}

func (f *FileWatcher) Watch() {
	go func() {
		for {
			time.Sleep(f.watchInv)

			func() {
				f.lock.Lock()
				defer f.lock.Unlock()

				for filepath := range f.files {
					newModtime, err := sys.Modtime(filepath)
					if err != nil {
						cluster := f.files[filepath]
						// Signal each connection to this file that the
						// file cannot be found.
						for _, conn := range cluster.conns {
							conn.Ch <- error_read
						}
						delete(f.files, filepath)
						log.Printf("Watch(): %s cannot be found\n", filepath)
						continue
					}

					cluster := f.files[filepath]
					if cluster.Lastmodifed != newModtime {
						fmt.Printf("%s was modified at: %s\n", filepath, time.Now().Local())

						// Update Lastmodifed time, otherwise it will be different each time.
						cluster.Lastmodifed = newModtime

						for _, conn := range cluster.conns {
							conn.Ch <- write_success
						}
					}
				}
			}()
		}
	}()
}

func NewFileWatcher(useHarness bool) *FileWatcher {
	watcher := &FileWatcher{
		files:    make(map[string]*connCluster),
		lock:     sync.Mutex{},
		watchInv: time.Duration(300 * time.Millisecond),
	}

	if useHarness {
		watcher.harness = NewTestHarness()
	}

	return watcher
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
