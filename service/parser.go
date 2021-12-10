package service

import (
	"bytes"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func convertMarkdownToHTML(pathToMarkdown string) ([]byte, error) {
	filedata, err := os.ReadFile(pathToMarkdown)
	if err != nil {
		return nil, err
	}

	// Add more parsing options
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.TaskList,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"), // Code highlight colors
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	var content bytes.Buffer
	if err := md.Convert(filedata, &content); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
