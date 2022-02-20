package testing

import (
	"bufio"
	"log"
	"os"
)

// To be used to deactivate timestamp logging during testing.
func NoTimestamp() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

func CaptureLog(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		return ""
	}
	log.SetOutput(writer)
	defer func() {
		reader.Close()
		writer.Close()
		log.SetOutput(os.Stderr)
	}()

	scanner := bufio.NewScanner(reader)
	go f()
	scanner.Scan()

	return scanner.Text()
}
