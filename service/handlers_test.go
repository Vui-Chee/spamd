package service

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"

	testtools "github.com/vui-chee/mdpreview/internal/testing"
)

var x embed.FS = testtools.MockFS

func init() {
	// Use this mock testing folder
	fsPrefix = "mockfs"
}

func TestGetEmbeddedCSS(t *testing.T) {
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
	// During testing, use this static testing folder instead.
	f = testtools.MockFS

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
    <div class="app">/README.md</div><div>light</div>
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
