package service

import (
	"embed"
	"net/http"
	"net/http/httptest"
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
