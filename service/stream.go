package service

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"spamd/internal/sys"
	"spamd/service/config"

	"github.com/gorilla/websocket"
)

const (
	ENDLESS_LOOP = -1

	// Messages for each connection.
	write_success = "success"
	error_read    = "error_read"
	close_conn    = "close"
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

// This interface is created following the API in
// gorilla/websocket. Only contains methods that will
// be used in the tool.
//
// This allows the connection to be mocked during
// testing.
type WebsocketConn interface {
	Close() error
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
}

type conn struct {
	// This channel acts as an event queue.
	// Any event that occurs on the WS connection may
	// trigger actions in other goroutines. So this
	// channel help facilitate that.
	Ch chan string

	// gorilla/websocket
	Conn WebsocketConn
}

func (c *conn) Trigger(event string) error {
	if event == write_success ||
		event == close_conn ||
		event == error_read {
		c.Ch <- event
		return nil
	}
	return fmt.Errorf("No such connection event.")
}

func NewConn(c WebsocketConn) *conn {
	return &conn{
		Ch:   make(chan string),
		Conn: c,
	}
}

func (c *conn) SendText(content []byte) error {
	return c.Conn.WriteMessage(websocket.TextMessage, content)
}

func (c *conn) SendConvertedMarkdownFromFile(filepath string) error {
	content, err := convertMarkdownToHTML(filepath)
	if err != nil {
		return err
	}

	err = c.SendText(content)
	if err != nil {
		return err
	}
	return nil
}

func (c *conn) OnReadConn(event string) (int, []byte, error) {
	ty, data, err := c.Conn.ReadMessage()

	// NOTE: someone must receive this otherwise, this will block.
	c.Trigger(event)

	if err != nil {
		return -1, nil, err
	}
	return ty, data, err
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

func (f *FileWatcher) AddConn(filepath string, modtime time.Time, c WebsocketConn) *conn {
	f.lock.Lock()
	defer f.lock.Unlock()

	newConn := NewConn(c)
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

	// Closes underlying network connection.
	cluster.conns[index].Conn.Close()

	updatedConnections := append(cluster.conns[:index], cluster.conns[index+1:]...)
	if len(updatedConnections) == 0 {
		// No more connections to this file, so drop key-value pair.
		delete(f.files, filepath)
	} else {
		cluster.conns = updatedConnections
	}

	return nil
}

func (f *FileWatcher) CloseClusterConn(filepath string) {
	cluster, ok := f.files[filepath]
	if ok {
		for _, c := range cluster.conns {
			c.Conn.Close()
		}
	}

	delete(f.files, filepath)
}

func (f *FileWatcher) CloseAllConn() {
	f.lock.Lock()
	defer f.lock.Unlock()

	// Delete all clusters.
	for filepath := range f.files {
		f.CloseClusterConn(filepath)
	}
}

func (f *FileWatcher) RefreshContent(w http.ResponseWriter, r *http.Request) {
	// Get the path relative to the directory where the tool is run.
	// '+1' to skip the leading '/'.
	uri := r.URL.Path
	filepath, err := filepath.EvalSymlinks(uri[len(config.RefreshPrefix)+1:])
	if err != nil {
		return
	}

	modtime, err := sys.Modtime(filepath)
	if err != nil {
		return
	}

	// Create new websocket connection.
	var upgrader = websocket.Upgrader{}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	// Add mapping storing the connection.
	conn := f.AddConn(filepath, modtime, wsConn)
	defer f.DeleteConn(filepath, conn) // Close() will be called here

	// Read first page
	if err := conn.SendConvertedMarkdownFromFile(filepath); err != nil {
		log.Fatalln(err)
		return
	}

	// Listen for close connection.
	//
	// NOTE:
	// You must listen inside another goroutine, since
	// ReadMessage() is blocking.
	go func() {
		for {
			conn.OnReadConn(close_conn)
		}
	}()

	for {
		// Used during testing only.
		if f.harness.loops > 0 {
			f.harness.loops--
		} else if f.harness.loops == 0 {
			break
		}

		select {
		case msg := <-conn.Ch:
			if msg == error_read {
				// During such error, file will be deleted from
				// trackFiles & messageChannels during Watch().
				log.Printf("Error watching file. %s is either deleted/renamed/moved.\n", filepath)
				return
			}

			if msg == write_success {
				if err := conn.SendConvertedMarkdownFromFile(filepath); err != nil {
					log.Fatalln(err)
					continue
				}
			}

			if msg == close_conn {
				log.Printf("Closed tab for %s\n", filepath)
				return
			}
		}
	}
}

func (f *FileWatcher) Watch() {
	go func() {
		// Only relevant during testing.
		if f.harness.useWaitGroup {
			f.harness.wg.Done()
		}

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
							conn.Trigger(error_read)
						}
						f.CloseClusterConn(filepath)
						log.Printf("Watch(): %s cannot be found\n", filepath)
						continue
					}

					cluster := f.files[filepath]
					if cluster.Lastmodifed != newModtime {
						fmt.Printf("%s was modified at: %s\n", filepath, time.Now().Local())

						// Update Lastmodifed time, otherwise it will be different each time.
						cluster.Lastmodifed = newModtime

						for _, conn := range cluster.conns {
							conn.Trigger(write_success)
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
