// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package schannel

import (
	"io"

	"github.com/alexbrainman/sspi"
)

type inputBuffer struct {
	data   []byte
	reader io.Reader
}

func newInputBuffer(initialsize int, reader io.Reader) *inputBuffer {
	return &inputBuffer{
		data:   make([]byte, 0, initialsize),
		reader: reader,
	}
}

// copy copies data d into buffer ib. copy grows destination if needed.
func (ib *inputBuffer) copy(d []byte) int {
	// TODO: check all call sites, maybe this can be made more efficient
	return copy(ib.data, d)
}

func (ib *inputBuffer) reset() {
	ib.data = ib.data[:0]
}

func (ib *inputBuffer) grow() {
	b := make([]byte, len(ib.data), cap(ib.data)*2)
	copy(b, ib.data)
	ib.data = b
}

func (ib *inputBuffer) readMore() error {
	if len(ib.data) == cap(ib.data) {
		ib.grow()
	}
	n0 := len(ib.data)
	ib.data = ib.data[:cap(ib.data)]
	n, err := ib.reader.Read(ib.data[n0:])
	if err != nil {
		return err
	}
	ib.data = ib.data[:n0+n]
	return nil
}

func (ib *inputBuffer) bytes() []byte {
	return ib.data
}

func sendOutBuffer(w io.Writer, b *sspi.SecBuffer) error {
	_, err := b.WriteAll(w)
	// TODO: see if I can preallocate buffers instead
	b.Free()
	b.Set(sspi.SECBUFFER_TOKEN, nil)
	return err
}

// indexOfSecBuffer searches buffers bs for buffer type buftype.
// It returns -1 if not found.
func indexOfSecBuffer(bs []sspi.SecBuffer, buftype uint32) int {
	for i := range bs {
		if bs[i].BufferType == buftype {
			return i
		}
	}
	return -1
}
