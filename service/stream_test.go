package service

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	conf "github.com/vui-chee/mdpreview/service/config"
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

func TestFormatStreamData(t *testing.T) {
	// Each data packet must end with newline.
	// A single newline is also a data packet - empty data packet.
	inputs := []string{
		"",
		"\n",
		"abc",
		"abc\n",
		"abc\ndef",
		"abc\ndef\n",

		// consecutive newlines
		"abc\n\n",
		"abc\n\ndef",
		"abc\n\n\ndef",
		"abc\n\ndef\n",
	}

	expected := []string{
		"",
		"data:\ndata:\n\n",
		"data:abc\n\n",
		"data:abc\ndata:\n\n",
		"data:abc\ndata:def\n\n",
		"data:abc\ndata:def\ndata:\n\n",

		"data:abc\ndata:\ndata:\n\n",
		"data:abc\ndata:\ndata:def\n\n",
		"data:abc\ndata:\ndata:\ndata:def\n\n",
		"data:abc\ndata:\ndata:def\ndata:\n\n",
	}

	for i, input := range inputs {
		got := eventStreamFormat(input)
		if !reflect.DeepEqual(got, []byte(expected[i])) {
			t.Errorf("case %d, eventStreamFormat returns \"%s\", expected \"%s\"", i+1, got, expected[i])
		}
	}
}

func TestErrorDuringRead(t *testing.T) {
	wantError := errors.New("Convert function fails.")

	file, _ := ioutil.TempFile(".", "*")
	converterMutex.Lock()
	savedConverter := converter
	converter = func(filedata []byte, content *bytes.Buffer) error {
		return wantError
	}
	defer func() {
		os.Remove(file.Name())
		converter = savedConverter
		converterMutex.Unlock()
	}()

	err := readAndSendMarkdown(nil, file.Name())
	if err != wantError {
		t.Errorf("got %s; want %s", err, wantError)
	}
}

func TestWriteContentAsDataPacket(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	file.WriteString("# Header")
	defer os.Remove(file.Name())

	rr := httptest.NewRecorder()
	readAndSendMarkdown(rr, file.Name())
	if !rr.Flushed {
		t.Error("Should have flushed.")
	}

	want := `data:<h1 id="header">Header</h1>
data:

`
	if rr.Body.String() != want {
		t.Errorf("got %s; want %s", rr.Body.String(), want)
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

	// file.Name() returns "./{uri}", skip first dot.
	resourceUri := conf.RefreshPrefix + file.Name()[1:]
	req, err := http.NewRequest("GET", resourceUri, nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(watcher.RefreshContent)

	handler.ServeHTTP(rr, req)

	// Read from byte stream.
	got := make([]byte, 200)
	_, err = rr.Result().Body.Read(got)
	if err != nil {
		t.Errorf("Error reading from event stream: %s", err)
	}

	want := `data:<h1 id="first-page">First Page</h1>
data:<p>An example tranformation of markdown contents into
data:actual HTML.</p>
data:<h2 id="xyz">XYZ</h2>
data:
`

	if !rr.Flushed {
		t.Error("Expected flushed.")
	}

	if string(got)[:len(want)] != want {
		t.Errorf("got %s; want %s", string(got), want)
	}
}

func TestCreateMappingWithInitialModtime(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	file.WriteString(`# First Page

An example tranformation of markdown contents into
actual HTML.

## Contents
`)
	defer os.Remove(file.Name())

	watcher := NewFileWatcher(true)
	watcher.harness.loops = 0 // Don't run main loop

	// file.Name() returns "./{uri}", skip first dot.
	resourceUri := conf.RefreshPrefix + file.Name()[1:]
	req, err := http.NewRequest("GET", resourceUri, nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(watcher.RefreshContent)

	handler.ServeHTTP(rr, req)

	// Check whether modtime matched
	for filepath := range watcher.files {
		osInfo, _ := os.Lstat(filepath)
		cluster := watcher.files[filepath]
		if osInfo.ModTime() != cluster.Lastmodifed {
			t.Errorf("got %s; want %s", cluster.Lastmodifed, osInfo.ModTime())
		}
	}
}

func TestCloseConnection(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	defer os.Remove(file.Name())

	watcher := NewFileWatcher(true)

	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	// Cancel request immdiately.
	cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", conf.RefreshPrefix+file.Name()[1:], nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(watcher.RefreshContent)

	handler.ServeHTTP(rr, req)

	select {
	case <-ctx.Done():
		// If success close.
	default:
		t.Error("Failed to close conn.")
	}
}

func TestAddNewConnection(t *testing.T) {
	watcher := NewFileWatcher(true)

	// 1. test add to empty set
	file1 := "first.txt"
	conn := watcher.AddConn(file1, time.Time{})
	if _, ok := watcher.files[file1]; !ok {
		t.Errorf("Connection for %s should be added. Got not found in map.", file1)
	}
	if conn == nil {
		t.Error("got <nil>; want *conn returned on AddConn()")
	}

	// 2. add to non-empty set
	watcher.AddConn(file1, time.Time{}) // add another connection to same file
	if cluster := watcher.files[file1]; len(cluster.conns) != 2 {
		t.Errorf("Want 2 connections under %s. Got %d.", file1, len(cluster.conns))
	}

	// 3. add to another empty set
	file2 := "second.txt"
	conn = watcher.AddConn(file2, time.Time{})
	cluster, ok := watcher.files[file2]
	if !ok {
		t.Errorf("Connection for %s should be added. Got non-found in map.", file2)
	}
	if len(cluster.conns) != 1 {
		t.Errorf("Want single connection under %s. Got %d.", file2, len(cluster.conns))
	}
}

func TestDeleteConnection(t *testing.T) {
	watcher := NewFileWatcher(true)
	file := "test.md"

	targetConn := &conn{
		Ch: make(chan string),
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
			watcher.AddConn(file, time.Time{})
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
