package service

import (
	"embed"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	testtools "github.com/vui-chee/mdpreview/internal/testing"
)

var (
	x embed.FS = testtools.MockFS

	// For locking shared FS during concurrent testing.
	// Although not required (since we are only dealing with concurrent reads)...
	// , you have to remind yourself in the future.
	fsMutex sync.Mutex
)

func init() {
	// Use this mock testing folder
	fsPrefix = "mockfs"
}

func TestGetEmbeddedCSS(t *testing.T) {
	fsMutex.Lock()
	defer fsMutex.Unlock()

	// During testing, use this static testing folder instead.
	f = testtools.MockFS

	req, err := http.NewRequest("GET", "/styles", nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(serveCSS)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("serveCSS returned wrong status code. Expected: %d. Got: %d.", http.StatusOK, status)
	}

	if got, want := rr.Header().Get("Content-Type"), "text/css"; got != want {
		t.Errorf("serveCSS returned wrong Content-Type. Expected %s. Got %s.", want, got)
	}

	if got, want := rr.Body.String(), `.app {
  background: green;
}
`; got != want {
		t.Errorf("serveCSS returned wrong body:\nExpected %s.\n--\nGot %s.", want, got)
	}
}

func TestServeCSS_ErrOnMissingFS(t *testing.T) {
	var fakeFS embed.FS
	// Change to non-existent folder
	f = fakeFS

	req, err := http.NewRequest("GET", "/styles", nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(serveCSS)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("serveCSS returned wrong status code. Expected: %d. Got: %d.", http.StatusOK, status)
	}
}

func TestGetEmbeddedHTML(t *testing.T) {
	fsMutex.Lock()
	defer fsMutex.Unlock()
	// During testing, use this static testing folder instead.
	f = testtools.MockFS

	// Mock Service Config.
	oldTheme := serviceConfig.Theme
	serviceConfig.SetTheme("light")        // set test theme
	defer serviceConfig.SetTheme(oldTheme) // Make sure to restore whatever theme.

	req, err := http.NewRequest("GET", "/README.md", nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(serveHTML)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("serveHTML returned wrong status code. Expected: %d. Got: %d.", http.StatusOK, status)
	}

	if got, want := rr.Header().Get("Content-Type"), "text/html"; got != want {
		t.Errorf("serveHTML returned wrong Content-Type. Expected %s. Got %s.", want, got)
	}

	if got, want := rr.Body.String(), `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>README.md</title>
  </head>
  <body>
    <div class="app">/README.md</div>
    <div>light</div>
    <div>/__/refresh</div>
    <div>/__/styles</div>
  </body>
</html>
`; got != want {
		t.Errorf("serveHTML returned wrong body:\nExpected %s.\n--\nGot %s.", want, got)
	}
}

func TestServeHTML_ErrOnMissingFS(t *testing.T) {
	var fakeFS embed.FS
	// Change to non-existent folder
	f = fakeFS

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(serveHTML)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("serveCSS returned wrong status code. Expected: %d. Got: %d.", http.StatusOK, status)
	}
}

func TestServeLocalImage(t *testing.T) {
	dir, err := ioutil.TempDir(".", "")
	if err != nil {
		t.Error("Failed to create tempdir.", err)
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	file, err := ioutil.TempFile(dir, "")
	if err != nil {
		t.Error("Failed to create tempfile.")
		t.FailNow()
	}
	fakeContents := "dummy-image-contents"
	file.Write([]byte(fakeContents))

	req, err := http.NewRequest("GET", file.Name(), nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(serveLocalImage)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status == http.StatusNotFound {
		t.Errorf("serveLocalImage returned 404.")
		t.FailNow()
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "image/" {
		t.Errorf("got %s; want %s\n", contentType, "image/")
		t.FailNow()
	}
	content := rr.Body.String()
	if content != fakeContents {
		t.Errorf("got %s; want %s\n", content, fakeContents)
		t.FailNow()
	}
}

func TestServeLocalImageNoSuchFile404(t *testing.T) {
	req, err := http.NewRequest("GET", "/no-such-file.png", nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(serveLocalImage)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("serveLocalImage returned %s.", rr.Result().Status)
		t.FailNow()
	}
}
