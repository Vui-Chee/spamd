package service

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

func TestErrorsOnAbsentFile(t *testing.T) {
	var err error

	filename := "fake-file"
	_, err = os.Lstat(filename)
	if err == nil {
		t.Error("fake-file should not exist.")
	}

	_, err = convertMarkdownToHTML(filename)
	if err == nil {
		t.Error("Should return error if file is absent.")
	}
}

func TestReturnErrorWhenConvertFails(t *testing.T) {
	wantError := errors.New("Convert function fails.")

	// Setup fake file and converter function.
	file, _ := ioutil.TempFile(".", "*")
	defer os.Remove(file.Name())
	savedConverter := converter
	defer func() {
		converter = savedConverter
	}()
	converter = func(filedata []byte, content *bytes.Buffer) error {
		return wantError
	}

	_, err := convertMarkdownToHTML(file.Name())
	if err == nil {
		t.Errorf("got <nil>; want error \"%s\"", wantError)
	}
}

func TestConvertIntoMarkdown(t *testing.T) {
	file, _ := ioutil.TempFile(".", "*")
	file.WriteString("# Header")
	defer os.Remove(file.Name())

	got, err := convertMarkdownToHTML(file.Name())
	if err != nil {
		t.Errorf("Should not return error. Got error \"%s\"", err)
	}

	want := `<h1 id="header">Header</h1>
`
	if string(got) != want {
		t.Errorf("got \"%s\"; want \"%s\"", string(got), want)
	}
}
