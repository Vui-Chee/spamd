package service

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	testutils "spamd/internal/testing"
	"spamd/service/config"
)

func TestConstructFileWatcher(t *testing.T) {
	var want string
	var got string

	watcher := NewFileWatcher(false)
	if watcher == nil {
		t.Error("got <nil>; want &FileWatcher{}")
	}

	want = "map[string]*service.connCluster"
	got = reflect.TypeOf(watcher.files).String()
	if got != want {
		t.Errorf("got %s; want %s", got, want)
	}

	want = "*sync.Mutex"
	got = reflect.TypeOf(&watcher.lock).String()
	if got != want {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestAddNewConnection(t *testing.T) {
	watcher := NewFileWatcher(true)
	// Create mock ws connection.
	c := &websocket.Conn{}

	// 1. test add to empty set
	file1 := "first.txt"
	conn := watcher.AddConn(file1, time.Time{}, c)
	if _, ok := watcher.files[file1]; !ok {
		t.Errorf("Connection for %s should be added. Got not found in map.", file1)
	}
	if conn == nil {
		t.Error("got <nil>; want *conn returned on AddConn()")
	}

	// 2. add to non-empty set
	watcher.AddConn(file1, time.Time{}, c) // add another connection to same file
	if cluster := watcher.files[file1]; len(cluster.conns) != 2 {
		t.Errorf("Want 2 connections under %s. Got %d.", file1, len(cluster.conns))
	}

	// 3. add to another empty set
	file2 := "second.txt"
	conn = watcher.AddConn(file2, time.Time{}, c)
	cluster, ok := watcher.files[file2]
	if !ok {
		t.Errorf("Connection for %s should be added. Got non-found in map.", file2)
	}
	if len(cluster.conns) != 1 {
		t.Errorf("Want single connection under %s. Got %d.", file2, len(cluster.conns))
	}
}

type MockWebsocketConn struct{}

func (c *MockWebsocketConn) Close() error {
	return nil
}

func (c *MockWebsocketConn) ReadMessage() (messageType int, p []byte, err error) {
	return 0, nil, nil
}

func (c *MockWebsocketConn) WriteMessage(messageType int, data []byte) error {
	return nil
}

func TestDeleteConnection(t *testing.T) {
	watcher := NewFileWatcher(true)
	file := "test.md"

	targetConn := &conn{
		Ch:   make(chan string),
		Conn: &MockWebsocketConn{},
	}

	// Initially no connections in map.
	watcher.files[file] = &connCluster{}

	// 1. try delete on empty set
	err := watcher.DeleteConn(file, targetConn)
	if err == nil {
		t.Error("got <nil>; want error.")
	}

	// 2. delete existing conn
	watcher.files[file] = &connCluster{
		conns: []*conn{
			targetConn,
		},
	}

	err = watcher.DeleteConn(file, targetConn)
	if err != nil {
		t.Errorf("got %s; want <nil>.", err)
	}
	if len(watcher.files) != 0 {
		t.Errorf("got %d; want 0.", len(watcher.files))
	}
}

func TestMultipleAddConn(t *testing.T) {
	var files = []string{
		"foo.md",
		"foo.md",
		"abc.md",
	}
	var wg sync.WaitGroup

	watcher := NewFileWatcher(true)
	for i := 0; i < len(files); i++ {
		file := files[i]
		wg.Add(1)
		go func() {
			watcher.AddConn(file, time.Time{}, &MockWebsocketConn{})
			wg.Done()
		}()
	}

	wg.Wait()

	// check contents of map
	if len(watcher.files) != 2 {
		t.Errorf("want 2 unique files; got %d", len(watcher.files))
	}
	if len(watcher.files["foo.md"].conns) != 2 {
		t.Errorf("want 2 connections under %s; got %d", "foo.md", len(watcher.files["foo.md"].conns))
	}
}

func createMockWsConn(resourceUri string, handler func(http.ResponseWriter, *http.Request)) (*httptest.Server, *websocket.Conn, error) {
	// Start a test server.
	s := httptest.NewServer(http.HandlerFunc(handler))

	// Connect to test server.
	u := "ws" + strings.TrimPrefix(s.URL, "http") + resourceUri
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		s.Close()
		return nil, nil, err
	}

	return s, ws, nil
}

func TestCloseConnection(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	defer os.Remove(file.Name())

	watcher := NewFileWatcher(true)
	resourceUri := config.RefreshPrefix + file.Name()[1:]

	s, _, err := createMockWsConn(resourceUri, watcher.RefreshContent)
	defer s.Close()
	if err != nil {
		t.Error(err)
	}

	filepath := file.Name()[2:]

	// Deactivate timestamp.
	testutils.NoTimestamp()

	cluster := watcher.files[filepath]
	for _, conn := range cluster.conns {
		conn.Trigger(close_conn)
	}

	want := close_conn
	for _, conn := range cluster.conns {
		got := <-conn.Ch
		if got != want {
			t.Errorf("got %s; want %s", got, want)
		}
	}
}

func TestGetFirstPageOnConnect(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	file.WriteString(`# First Page

An example tranformation of markdown contents into
actual HTML.

## XYZ
`)
	defer os.Remove(file.Name())

	watcher := NewFileWatcher(true)
	watcher.harness.loops = 0 // Don't run main loop
	resourceUri := config.RefreshPrefix + file.Name()[1:]

	s, ws, err := createMockWsConn(resourceUri, watcher.RefreshContent)
	defer s.Close()
	if err != nil {
		t.Error(err)
	}

	// Read from websocket.
	_, got, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("Error reading websocket connection: %s", err)
	}

	want := `<h1 id="first-page">First Page</h1>
<p>An example tranformation of markdown contents into
actual HTML.</p>
<h2 id="xyz">XYZ</h2>
`
	if string(got) != want {
		t.Errorf("got %s; want %s", string(got), want)
	}
}

func TestTriggerWriteOnWatch(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	defer os.Remove(file.Name())
	info, _ := file.Stat()
	filepath := file.Name()[2:]

	watcher := NewFileWatcher(true)
	// Setup fake conn struct.
	watcher.files[filepath] = &connCluster{
		Lastmodifed: info.ModTime(),
		conns: []*conn{
			{
				Ch: make(chan string),
			},
		},
	}

	watcher.harness.useWaitGroup = true
	watcher.watchInv = 0 // No delay between each loop.
	watcher.harness.wg.Add(1)
	watcher.Watch()
	watcher.harness.wg.Wait() // wait for goroutine to start.

	// Space out the write time, otherwise, the difference in
	// time may be negligible.
	time.Sleep(30 * time.Millisecond)
	file.WriteString("Next paragraph.")

	// Check if channel is written to with correct message.
	for _, conn := range watcher.files[filepath].conns {
		msg := <-conn.Ch
		if msg != write_success {
			t.Errorf("got %s; want %s", msg, write_success)
		}
	}
}

func TestTriggerErrorOnWatch(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	file.WriteString("# First Page")
	info, _ := file.Stat()
	filepath := file.Name()[2:]

	watcher := NewFileWatcher(true)
	// Setup fake conn struct.
	watcher.files[filepath] = &connCluster{
		Lastmodifed: info.ModTime(),
		conns: []*conn{
			{
				Ch:   make(chan string),
				Conn: &MockWebsocketConn{},
			},
		},
	}
	watcher.harness.useWaitGroup = true
	watcher.watchInv = 0 // No delay between each loop.
	watcher.harness.wg.Add(1)
	watcher.Watch()
	watcher.harness.wg.Wait() // wait for goroutine to start.

	// Now drop file, should trigger err during sys.Modtime
	time.Sleep(30 * time.Millisecond)
	os.Remove(file.Name())

	for _, conn := range watcher.files[filepath].conns {
		msg := <-conn.Ch
		if msg != error_read {
			t.Errorf("got %s; want %s", msg, error_read)
		}
	}
}
