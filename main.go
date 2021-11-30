package main

import (
	"fmt"
	"log"
	"os"

	"bytes"

	"github.com/yuin/goldmark"
)

// TODO
// create simple server, serving generated html from README.
func main() {
	dat, err := os.ReadFile("hello.md")
	if err != nil {
		log.Fatalln("Error reading file.")
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(dat, &buf); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}
