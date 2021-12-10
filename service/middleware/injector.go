package middleware

import (
	"context"
	"net/http"
)

// Use this pass needed commandline arguments into
// request context.
type Args struct {
	Filepath string
}

// Middleware to inject commandline args into request context.
type ArgsInjector struct {
	handler http.Handler

	args Args
}

func (l *ArgsInjector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "filepath", l.args.Filepath)
	updatedRequest := r.WithContext(ctx)
	l.handler.ServeHTTP(w, updatedRequest)
}

func NewArgsInjector(handler http.Handler, args Args) *ArgsInjector {
	return &ArgsInjector{handler, args}
}
