package gitmedia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
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
	header := fmt.Sprintf("# %d\n", len(MediaWarning))
	e.writer.Write([]byte(header))
	e.writer.Write(MediaWarning)
	return e.jsonencoder.Encode(obj)
}

type Decoder struct {
	reader      io.Reader
	jsondecoder *json.Decoder
}

func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{reader, json.NewDecoder(reader)}
}

func (d *Decoder) Decode(obj interface{}) error {
	buf := make([]byte, 10)
	d.reader.Read(buf)
	slices := bytes.SplitN(buf, []byte("\n"), 2)
	headerlen, err := strconv.Atoi(string(slices[0])[2:])
	if err != nil {
		fmt.Printf("Error reading header:\n%s\n", string(buf))
		panic(err)
	}

	buf = make([]byte, headerlen-len(slices[1]))
	d.reader.Read(buf)

	return d.jsondecoder.Decode(obj)
}
