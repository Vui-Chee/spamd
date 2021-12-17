package service

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:embed mockfs
var mockFS embed.FS

func init() {
	// Use this mock testing folder
	fsPrefix = "mockfs"
}

func TestServeCSS(t *testing.T) {
	// During testing, use this static testing folder instead.
	f = mockFS

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

	gotContentType := rr.Header().Get("Content-Type")
	wantContentType := "text/css"
	if gotContentType != wantContentType {
		t.Errorf("serveCSS returned wrong Content-Type. Expected %s. Got %s.", wantContentType, gotContentType)
	}

	gotBody := rr.Body.String()
	wantBody := `.app {
  background: green;
}
`
	if gotBody != wantBody {
		t.Errorf("serveCSS returned wrong body:\nExpected %s.\n--\nGot %s.", wantBody, gotBody)
	}
}

func TestServeCSSwithError(t *testing.T) {
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
