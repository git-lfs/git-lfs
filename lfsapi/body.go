package lfsapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type ReadSeekCloser interface {
	io.Seeker
	io.ReadCloser
}

func MarshalToRequest(req *http.Request, obj interface{}) error {
	by, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	req.ContentLength = int64(len(by))
	req.Body = NewByteBody(by)
	return nil
}

func NewByteBody(by []byte) ReadSeekCloser {
	return &closingByteReader{Reader: bytes.NewReader(by)}
}

type closingByteReader struct {
	*bytes.Reader
}

func (r *closingByteReader) Close() error {
	return nil
}
