package service

import (
	"embed"
	"fmt"
	"os"
	"strings"
	"time"
)

//go:embed build/styles.css
//go:embed build/index.html
var f embed.FS

func exitOnError(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	os.Exit(1)
}

func getEmbeddedBytes(filepath string) ([]byte, error) {
	data, err := f.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getFileModtime(filename string) (time.Time, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return time.Time{}, err
	}

	return info.ModTime(), nil
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
	var eventPayload string
	dataLines := strings.Split(data, "\n")
	for _, line := range dataLines {
		eventPayload = eventPayload + "data: " + line + "\n"
	}
	return []byte(eventPayload + "\n")
}
