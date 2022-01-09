package testing

import (
	"bytes"
	"log"
	"os"
)

// To be used to deactivate timestamp logging during testing.
func NoTimestamp() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

func CaptureLog(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}
