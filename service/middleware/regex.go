package middleware

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/vui-chee/spamd/internal/sys"
)

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route

	AdditionalCheck func(path string) bool
}

func (h *RegexpHandler) Handler(pattern string, handler http.Handler) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		sys.ErrorAndExit(fmt.Sprintf("Failed to compile regex for %s.", pattern))
	}

	h.routes = append(h.routes, &route{regex, handler})
}

func (h *RegexpHandler) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		sys.ErrorAndExit(fmt.Sprintf("Failed to compile regex for %s.", pattern))
	}

	h.routes = append(h.routes, &route{regex, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) && h.AdditionalCheck(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}
