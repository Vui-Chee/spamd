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

	want = "map[string]*service.fileInfo"
	got = reflect.TypeOf(watcher.trackFiles).String()
	if got != want {
		t.Errorf("got %s; want %s", got, want)
	}

	want = "map[chan string]string"
	got = reflect.TypeOf(watcher.messageChannels).String()
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
	for filepath, info := range watcher.trackFiles {
		osInfo, _ := os.Lstat(filepath)
		if osInfo.ModTime() != info.Lastmodifed {
			t.Errorf("got %s; want %s", info.Lastmodifed, osInfo.ModTime())
		}
	}
}

func TestWatcherTriggersChannelOnWrite(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	info, _ := os.Lstat(file.Name())
	defer os.Remove(file.Name())

	// Must be a pointer since it's passed into a
	// function inside Watch().
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	wg.Add(1)

	watcher := NewFileWatcher(true)
	// messageChannels
	testChannel := make(chan string)
	watcher.messageChannels[testChannel] = file.Name()
	// trackFiles
	watcher.trackFiles[file.Name()] = &fileInfo{
		Count:       1,
		Lastmodifed: info.ModTime(),
	}

	watcher.Watch(func() {
		wg.Done()
	}) // runs a goroutine

	// Now write to file.
	file.WriteString("Write content to file")
	wg.Wait()

	latest, _ := os.Lstat(file.Name())
	want := latest.ModTime().String()
	got := <-testChannel
	if got != want {
		t.Errorf("got %s; want %s", got, want)
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
