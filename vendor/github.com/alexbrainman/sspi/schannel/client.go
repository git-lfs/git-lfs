// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package schannel

import (
	"errors"
	"io"
	"syscall"
	"unsafe"

	"github.com/alexbrainman/sspi"
)

// TODO: add documentation

// TODO: maybe come up with a better name

type Client struct {
	ctx   *sspi.Context
	conn  io.ReadWriter
	inbuf *inputBuffer
}

func NewClientContext(cred *sspi.Credentials, conn io.ReadWriter) *Client {
	return &Client{
		ctx:  sspi.NewClientContext(cred, sspi.ISC_REQ_STREAM|sspi.ISC_REQ_ALLOCATE_MEMORY|sspi.ISC_REQ_EXTENDED_ERROR|sspi.ISC_REQ_MANUAL_CRED_VALIDATION),
		conn: conn,
		// TODO: decide how large this buffer needs to be (it cannot be too small otherwise messages won't fit)
		inbuf: newInputBuffer(1000, conn),
	}
}

func (c *Client) Handshake(serverName string) error {
	name, err := syscall.UTF16PtrFromString(serverName)
	if err != nil {
		return err
	}
	inBuf := []sspi.SecBuffer{
		{BufferType: sspi.SECBUFFER_TOKEN},
		{BufferType: sspi.SECBUFFER_EMPTY},
	}
	// TODO: InitializeSecurityContext doco says that inBufs should be nil on the first call
	inBufs := sspi.NewSecBufferDesc(inBuf[:])
	outBuf := []sspi.SecBuffer{
		{BufferType: sspi.SECBUFFER_TOKEN},
	}
	outBufs := sspi.NewSecBufferDesc(outBuf)

	for {
		ret := c.ctx.Update(name, outBufs, inBufs)

		// send data to peer
		err := sendOutBuffer(c.conn, &outBuf[0])
		if err != nil {
			return err
		}

		// update input buffer
		fetchMore := true
		switch ret {
		case sspi.SEC_E_OK, sspi.SEC_I_CONTINUE_NEEDED:
			if inBuf[1].BufferType == sspi.SECBUFFER_EXTRA {
				c.inbuf.copy(inBuf[1].Bytes())
				fetchMore = false
			} else {
				c.inbuf.reset()
			}
		}

		// decide what to do next
		switch ret {
		case sspi.SEC_E_OK:
			// negotiation is competed
			return nil
		case sspi.SEC_I_CONTINUE_NEEDED, sspi.SEC_E_INCOMPLETE_MESSAGE:
			// continue on
		default:
			return ret
		}

		// fetch more input data if needed
		if fetchMore {
			err := c.inbuf.readMore()
			if err != nil {
				return err
			}
		}
		inBuf[0].Set(sspi.SECBUFFER_TOKEN, c.inbuf.bytes())
		inBuf[1].Set(sspi.SECBUFFER_EMPTY, nil)
	}
}

// TODO: protect Handshake, Read, Write and Shutdown with locks
// TODO: call Handshake at the start Read and Write unless handshake is already complete

func (c *Client) writeBlock(data []byte) (int, error) {
	ss, err := c.streamSizes()
	if err != nil {
		return 0, err
	}
	// TODO: maybe make this buffer (and header and trailer buffers) part of Context struct
	var b [4]sspi.SecBuffer
	b[0].Set(sspi.SECBUFFER_STREAM_HEADER, make([]byte, ss.Header))
	b[1].Set(sspi.SECBUFFER_DATA, data)
	b[2].Set(sspi.SECBUFFER_STREAM_TRAILER, make([]byte, ss.Trailer))
	b[3].Set(sspi.SECBUFFER_EMPTY, nil)
	ret := sspi.EncryptMessage(c.ctx.Handle, 0, sspi.NewSecBufferDesc(b[:]), 0)
	switch ret {
	case sspi.SEC_E_OK:
	case sspi.SEC_E_CONTEXT_EXPIRED:
		// TODO: handle this
		panic("writeBlock: SEC_E_CONTEXT_EXPIRED")
	default:
		return 0, ret
	}
	n1, err := b[0].WriteAll(c.conn)
	if err != nil {
		return n1, err
	}
	n2, err := b[1].WriteAll(c.conn)
	if err != nil {
		return n1 + n2, err
	}
	n3, err := b[2].WriteAll(c.conn)
	return n1 + n2 + n3, err
}

func (c *Client) Write(b []byte) (int, error) {
	ss, err := c.streamSizes()
	if err != nil {
		return 0, err
	}
	// TODO: handle redoing context here
	total := 0
	for len(b) > 0 {
		// TODO: maybe use ss.BlockSize to decide on optimum block size
		b2 := b
		if len(b) > int(ss.MaximumMessage) {
			b2 = b2[:ss.MaximumMessage]
		}
		n, err := c.writeBlock(b2)
		total += n
		if err != nil {
			return total, err
		}
		b = b[len(b2):]
	}
	return total, nil
}

func (c *Client) Read(data []byte) (int, error) {
	if len(c.inbuf.bytes()) == 0 {
		err := c.inbuf.readMore()
		if err != nil {
			return 0, err
		}
	}
	var b [4]sspi.SecBuffer
	desc := sspi.NewSecBufferDesc(b[:])
loop:
	for {
		b[0].Set(sspi.SECBUFFER_DATA, c.inbuf.bytes())
		b[1].Set(sspi.SECBUFFER_EMPTY, nil)
		b[2].Set(sspi.SECBUFFER_EMPTY, nil)
		b[3].Set(sspi.SECBUFFER_EMPTY, nil)
		ret := sspi.DecryptMessage(c.ctx.Handle, desc, 0, nil)
		switch ret {
		case sspi.SEC_E_OK:
			break loop
		case sspi.SEC_E_INCOMPLETE_MESSAGE:
			// TODO: it seems b[0].BufferSize or b[1].BufferSize contains "how many more bytes needed for full message" - maybe use it somehow
			// read more and try again
			err := c.inbuf.readMore()
			if err != nil {
				return 0, err
			}
		default:
			// TODO: handle other ret values
			return 0, errors.New("not implemented")
		}
	}
	i := indexOfSecBuffer(b[:], sspi.SECBUFFER_DATA)
	if i == -1 {
		return 0, errors.New("DecryptMessage did not return SECBUFFER_DATA")
	}
	n := copy(data, b[i].Bytes())
	i = indexOfSecBuffer(b[:], sspi.SECBUFFER_EXTRA)
	if i == -1 {
		c.inbuf.reset()
	} else {
		c.inbuf.copy(b[i].Bytes())
	}
	return n, nil
}

func (c *Client) applyShutdownControlToken() error {
	data := uint32(_SCHANNEL_SHUTDOWN)
	b := sspi.SecBuffer{
		BufferType: sspi.SECBUFFER_TOKEN,
		Buffer:     (*byte)(unsafe.Pointer(&data)),
		BufferSize: uint32(unsafe.Sizeof(data)),
	}
	desc := sspi.SecBufferDesc{
		Version:      sspi.SECBUFFER_VERSION,
		BuffersCount: 1,
		Buffers:      &b,
	}
	ret := sspi.ApplyControlToken(c.ctx.Handle, &desc)
	if ret != sspi.SEC_E_OK {
		return ret
	}
	return nil
}

func (c *Client) Shutdown() error {
	err := c.applyShutdownControlToken()
	if err != nil {
		return err
	}
	inBuf := []sspi.SecBuffer{
		{BufferType: sspi.SECBUFFER_TOKEN},
		{BufferType: sspi.SECBUFFER_EMPTY},
	}
	inBufs := sspi.NewSecBufferDesc(inBuf[:])
	outBuf := []sspi.SecBuffer{
		{BufferType: sspi.SECBUFFER_TOKEN},
	}
	outBufs := sspi.NewSecBufferDesc(outBuf)
	for {
		// TODO: I am not sure if I can pass nil as targname
		ret := c.ctx.Update(nil, outBufs, inBufs)

		// send data to peer
		err := sendOutBuffer(c.conn, &outBuf[0])
		if err != nil {
			return err
		}

		// update input buffer
		fetchMore := true
		switch ret {
		case sspi.SEC_E_OK, sspi.SEC_I_CONTINUE_NEEDED:
			if inBuf[1].BufferType == sspi.SECBUFFER_EXTRA {
				c.inbuf.copy(inBuf[1].Bytes())
				fetchMore = false
			} else {
				c.inbuf.reset()
			}
		}

		// decide what to do next
		switch ret {
		case sspi.SEC_E_OK, sspi.SEC_E_CONTEXT_EXPIRED:
			// shutdown is competed
			return nil
		case sspi.SEC_I_CONTINUE_NEEDED, sspi.SEC_E_INCOMPLETE_MESSAGE:
			// continue on
		default:
			return ret
		}

		// fetch more input data if needed
		if fetchMore {
			err := c.inbuf.readMore()
			if err != nil {
				return err
			}
		}
		inBuf[0].Set(sspi.SECBUFFER_TOKEN, c.inbuf.bytes())
		inBuf[1].Set(sspi.SECBUFFER_EMPTY, nil)
	}
}
