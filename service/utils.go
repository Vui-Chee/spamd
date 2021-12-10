package service

import (
	"embed"
	"fmt"
	"os"
)

//go:embed frontend
var f embed.FS

func exitOnError(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	os.Exit(1)
}

func getEmbeddedBytes(filepath string) ([]byte, error) {
	data, err := f.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
