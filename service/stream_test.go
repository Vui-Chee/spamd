package service

import (
	"bytes"
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

	watcher := NewFileWatcher()
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

## Contents
`)
	defer os.Remove(file.Name())

	watcher := NewFileWatcher()
	watcher.loops = 1

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
data:<h2 id="contents">Contents</h2>
data:

`

	if !rr.Flushed {
		t.Error("Expected flushed.")
	}

	if string(got)[:len(want)] != want {
		t.Errorf("got %s; want %s", string(got), want)
	}
}
