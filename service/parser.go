package service

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	converterMutex sync.Mutex

	// A function that transforms a sequence of bytes into
	// markdown content.
	converter = func(filedata []byte, content *bytes.Buffer) error {
		md := goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.TaskList,
				highlighting.NewHighlighting(
					highlighting.WithStyle(serviceConfig.CodeBlockTheme), // Code highlight colors
				),
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
			),
		)

		return md.Convert(filedata, content)
	}
)

func convertMarkdownToHTML(pathToMarkdown string) ([]byte, error) {
	filedata, err := os.ReadFile(pathToMarkdown)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %s", pathToMarkdown, err)
	}

	var content bytes.Buffer
	if err := converter(filedata, &content); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
