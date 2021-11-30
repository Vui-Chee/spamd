package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

// TODO
// create simple server, serving generated html from README.
func main() {
	dat, err := os.ReadFile("hello.md")
	if err != nil {
		log.Fatalln("Error reading file.")
		os.Exit(1)
	}

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	html := string(markdown.ToHTML(dat, parser, nil))

	fmt.Println(html)
}
