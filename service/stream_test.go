package service

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
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
