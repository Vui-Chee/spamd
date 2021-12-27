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

	// file.Name() returns "./{uri}", skip first dot.
	resourceUri := conf.RefreshPrefix + file.Name()[1:]
	req, err := http.NewRequest("GET", resourceUri, nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(watcher.RefreshContent)

	go func() {
		for channel := range watcher.messageChannels {
			channel <- time.Now().String()
		}
	}()

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

func TestUpdateContentOnWrite(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	file.WriteString(`# First Page

An example tranformation of markdown contents into
actual HTML.

## Contents
`)

	defer os.Remove(file.Name())

	watcher := NewFileWatcher(true)
	watcher.harness.useWaitGroup = true
	// +1 wait for mapping to be constructed based on initial
	// modtime.
	watcher.harness.Incr()

	// file.Name() returns "./{uri}", skip first dot.
	resourceUri := conf.RefreshPrefix + file.Name()[1:]
	req, err := http.NewRequest("GET", resourceUri, nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(watcher.RefreshContent)

	// Run initial request first to construct the
	// mapping based on the file's initial modtime.
	go func() {
		handler.ServeHTTP(rr, req)
	}()

	// Wait for mapping to be constructed first.
	watcher.harness.wg.Wait()

	// Now write to file BEFORE watcher.
	//
	// File modtime will change. And then watcher will trigger the
	// channel within the handler to output the new contents.
	file.WriteString("### new content")

	// Start watching the files in the current directory.
	// In this case, the service folder.
	//
	// Ensure the code runs once after writing to file.
	// This goroutine runs every X seconds. So there may be cases
	// where the code does not get run.
	watcher.harness.Incr() // +1 for watcher to trigger
	watcher.harness.Incr() // +1 for handler to read new changes
	watcher.Watch()        // Decr() is called after main code

	// Now wait for both goroutines to finish their tasks.
	watcher.harness.wg.Wait()

	// The first write happens after first connection opens.
	// The second write occurs when the watcher reads the file
	// change and triggers the next read.
	want := `data:<h1 id="first-page">First Page</h1>
data:<p>An example tranformation of markdown contents into
data:actual HTML.</p>
data:<h2 id="contents">Contents</h2>
data:

data:<h1 id="first-page">First Page</h1>
data:<p>An example tranformation of markdown contents into
data:actual HTML.</p>
data:<h2 id="contents">Contents</h2>
data:<h3 id="new-content">new content</h3>
data:

`
	// Read from byte stream.
	got := make([]byte, len(want))
	_, err = rr.Result().Body.Read(got)
	if err != nil {
		t.Errorf("Error reading from event stream: %s", err)
	}

	if !rr.Flushed {
		t.Error("Expected flushed.")
	}

	if string(got[:len(want)]) != want {
		t.Errorf("got %s; want %s", string(got), want)
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
