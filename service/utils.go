package service

import (
	"fmt"
	"os"
)

func exitOnError(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
	}
	os.Exit(1)
}
