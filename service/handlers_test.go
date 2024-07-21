package service

import (
	"embed"
	"net/http"
	"os"
	"sync"
	"testing"

	testtools "spamd/internal/testing"
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

	rr := testtools.MockRequest(t,
		"GET",
		"/styles",
		http.HandlerFunc(serveCSS),
	)

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

	rr := testtools.MockRequest(t,
		"GET",
		"/styles",
		http.HandlerFunc(serveCSS),
	)

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

	rr := testtools.MockRequest(t,
		"GET",
		"/README.md",
		http.HandlerFunc(serveHTML),
	)

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

	rr := testtools.MockRequest(t,
		"GET",
		"/",
		http.HandlerFunc(serveHTML),
	)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("serveCSS returned wrong status code. Expected: %d. Got: %d.", http.StatusOK, status)
	}
}

func TestServeLocalImage(t *testing.T) {
	dir, err := os.MkdirTemp(".", "")
	if err != nil {
		t.Error("Failed to create tempdir.", err)
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	file, err := os.CreateTemp(dir, "")
	if err != nil {
		t.Error("Failed to create tempfile.")
		t.FailNow()
	}
	fakeContents := "dummy-image-contents"
	file.Write([]byte(fakeContents))

	rr := testtools.MockRequest(t,
		"GET",
		file.Name(),
		http.HandlerFunc(serveLocalImage),
	)

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
	rr := testtools.MockRequest(t,
		"GET",
		"/no-such-file.png",
		http.HandlerFunc(serveLocalImage),
	)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("serveLocalImage returned %s.", rr.Result().Status)
		t.FailNow()
	}
}

func TestGetSvgImage(t *testing.T) {
	file, _ := os.Open("../assets/android.svg")

	rr := testtools.MockRequest(t,
		"GET",
		file.Name(),
		http.HandlerFunc(serveLocalImage),
	)

	got := rr.Header().Get("Content-Type")
	want := "image/svg+xml"
	if got != want {
		t.Errorf("got %s; want %s\n", got, want)
		t.FailNow()
	}
}
