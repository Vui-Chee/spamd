package service

import (
	"embed"
)

type Embedder struct {
	fs embed.FS

	store map[string]([]byte)
}

func (e *Embedder) GetBytes(filepath string) ([]byte, error) {
	data, ok := e.store[filepath]

	// Cache the bytes
	if !ok {
		data, err := e.fs.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
		e.store[filepath] = data
	}

	return data, nil
}
