package service

import (
	"embed"
	"os"
	"time"
)

//go:embed build/styles.css
//go:embed build/index.html
var f embed.FS

func getEmbeddedBytes(filepath string) ([]byte, error) {
	data, err := f.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getFileModtime(filename string) (time.Time, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return time.Time{}, err
	}

	return info.ModTime(), nil
}
