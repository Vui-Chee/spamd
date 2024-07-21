package service

import (
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"

	"spamd/service/config"
)

var (
	// Tests are executed by goroutines and therefore should assume
	// that concurrency concerns may arise.
	confMu sync.Mutex
)

func TestListenReturnsErrOnInvalidPort(t *testing.T) {
	invalidPorts := []int{
		-1,
		1,
		1023,
	}

	for _, port := range invalidPorts {
		_, err := listen(port)
		if err == nil {
			t.Errorf("Should return error if port == %d, Got: error == nil.\n", port)
		}
	}
}

func TestListenOnConfigPortOnZeroPort(t *testing.T) {
	confMu.Lock()
	// Should default to serviceConfig port (if non-zero).
	savedPort := serviceConfig.Port
	defer func() {
		serviceConfig.Port = savedPort
		confMu.Unlock()
	}()

	wantPort := 5817
	serviceConfig.Port = wantPort

	l, _ := listen(0)
	gotPort, _ := strconv.Atoi(strings.SplitAfter(l.Addr().String(), ":")[1])
	if gotPort != wantPort {
		t.Errorf("service.Listen(0): want %d, got: %d\n", wantPort, gotPort)
	}
}

func TestValidRedirects(t *testing.T) {
	var got bool

	got = redirectIfNotMarkdown(config.StylesPrefix)
	if got != true {
		t.Errorf("redirectIfNotMarkdown(\"%s\") should return true.", config.StylesPrefix)
	}

	file, _ := os.CreateTemp(".", "*.md")
	defer os.Remove(file.Name())

	uri := "/" + path.Base(file.Name())
	got = redirectIfNotMarkdown(uri)
	if got != true {
		t.Errorf("redirectIfNotMarkdown(\"%s\") should return true.", uri)
	}

	// Test serving resource from refresh route.
	uri = config.RefreshPrefix + "/" + path.Base(file.Name())
	got = redirectIfNotMarkdown(uri)
	if got != true {
		t.Errorf("redirectIfNotMarkdown(\"%s\") should return true.", uri)
	}
}

func TestRedirectOnNoSuchFile(t *testing.T) {
	uri := "/file-no-exists.md"
	got := redirectIfNotMarkdown(uri)
	if got != false {
		t.Errorf("redirectIfNotMarkdown(\"%s\") should return false.", uri)
	}
}

func TestOverrideTheme(t *testing.T) {
	confMu.Lock()
	savedTheme := serviceConfig.Theme
	savedCodeStyle := serviceConfig.CodeBlockTheme
	defer func() {
		// reset configs
		serviceConfig.SetTheme(savedTheme)
		serviceConfig.SetCodeBlockTheme(savedCodeStyle)
		confMu.Unlock()
	}()

	wantTheme := "dark"
	wantCodestyle := "xcode"

	overrideConfig(wantTheme, wantCodestyle)
	if serviceConfig.Theme != wantTheme {
		t.Errorf("OverrideConfig() : want %s, got %s\n", wantTheme, serviceConfig.Theme)
	}

	if serviceConfig.CodeBlockTheme != wantCodestyle {
		t.Errorf("OverrideConfig() : want %s, got %s\n", wantCodestyle, serviceConfig.CodeBlockTheme)
	}

}
