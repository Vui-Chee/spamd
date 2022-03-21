package testing

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func MockRequest(t *testing.T, method string, url string, handler http.Handler) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}
