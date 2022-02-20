package testing

import (
	"log"
	"testing"
)

const (
	test_msg = "Testing123"
)

func logSomething() {
	log.Print(test_msg)
}

func TestCaptureLog(t *testing.T) {
	NoTimestamp()

	want := test_msg
	got := CaptureLog(logSomething)
	if got != want {
		t.Errorf("got %s; want %s\n", got, want)
		t.FailNow()
	}
}
