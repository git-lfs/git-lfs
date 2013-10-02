package gitmedia

import (
	"bytes"
	"encoding/json"
	"io"
)

var MediaWarning = []byte("# This is a placeholder for large media, please install GitHub git-media to retrieve content\n# It is also possible you did not have the media locally, run 'git media sync' to retrieve it\n")

type Encoder struct {
	writer      io.Writer
	jsonencoder *json.Encoder
}

func NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{writer, json.NewEncoder(writer)}
}

func (e *Encoder) Encode(obj interface{}) error {
	e.writer.Write(MediaWarning)
	return e.jsonencoder.Encode(obj)
}

type Decoder struct {
	reader io.Reader
}

func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{reader}
}

func (d *Decoder) Decode(obj interface{}) error {
	buf := make([]byte, 1024)
	io.ReadFull(d.reader, buf)
	slices := bytes.Split(buf, []byte("\n"))
	dec := json.NewDecoder(bytes.NewBuffer(slices[len(slices)-2]))
	return dec.Decode(obj)
}
