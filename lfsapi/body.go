package lfsapi

import (
	"bytes"
	"encoding/json"
	"io"
)

type ReadSeekCloser interface {
	io.Seeker
	io.ReadCloser
}

func Marshal(obj interface{}) (ReadSeekCloser, error) {
	by, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return &closingByteReader{Reader: bytes.NewReader(by)}, nil
}

type closingByteReader struct {
	*bytes.Reader
}

func (r *closingByteReader) Close() error {
	return nil
}
