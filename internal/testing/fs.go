package testing

import (
	"embed"
)

// This mocks the embedded fs used in the static handlers.
//go:embed mockfs
var MockFS embed.FS
