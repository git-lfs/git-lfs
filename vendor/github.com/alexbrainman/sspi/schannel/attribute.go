// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package schannel

import (
	"syscall"
	"unsafe"

	"github.com/alexbrainman/sspi"
)

// TODO: maybe move all these into a separate package or something

func (c *Client) streamSizes() (*_SecPkgContext_StreamSizes, error) {
	// TODO: do not retrive _SecPkgContext_StreamSizes every time (cache the data and invalidate it every time is possible can be changed: handshake, redo, ...)
	// TODO: maybe return (header, trailer, maxmsg int, err error) instead
	// TODO: maybe this needs to be exported
	var ss _SecPkgContext_StreamSizes
	ret := sspi.QueryContextAttributes(c.ctx.Handle, _SECPKG_ATTR_STREAM_SIZES, (*byte)(unsafe.Pointer(&ss)))
	if ret != sspi.SEC_E_OK {
		return nil, ret
	}
	return &ss, nil
}

func (c *Client) ProtocolInfo() (name string, major, minor uint32, err error) {
	var pi _SecPkgContext_ProtoInfo
	ret := sspi.QueryContextAttributes(c.ctx.Handle, _SECPKG_ATTR_PROTO_INFO, (*byte)(unsafe.Pointer(&pi)))
	if ret != sspi.SEC_E_OK {
		return "", 0, 0, ret
	}
	defer sspi.FreeContextBuffer((*byte)(unsafe.Pointer(pi.ProtocolName)))
	s := syscall.UTF16ToString((*[2 << 20]uint16)(unsafe.Pointer(pi.ProtocolName))[:])
	return s, pi.MajorVersion, pi.MinorVersion, nil
}

func (c *Client) UserName() (string, error) {
	var ns _SecPkgContext_Names
	ret := sspi.QueryContextAttributes(c.ctx.Handle, _SECPKG_ATTR_NAMES, (*byte)(unsafe.Pointer(&ns)))
	if ret != sspi.SEC_E_OK {
		return "", ret
	}
	defer sspi.FreeContextBuffer((*byte)(unsafe.Pointer(ns.UserName)))
	s := syscall.UTF16ToString((*[2 << 20]uint16)(unsafe.Pointer(ns.UserName))[:])
	return s, nil
}

func (c *Client) AuthorityName() (string, error) {
	var a _SecPkgContext_Authority
	ret := sspi.QueryContextAttributes(c.ctx.Handle, _SECPKG_ATTR_AUTHORITY, (*byte)(unsafe.Pointer(&a)))
	if ret != sspi.SEC_E_OK {
		return "", ret
	}
	defer sspi.FreeContextBuffer((*byte)(unsafe.Pointer(a.AuthorityName)))
	s := syscall.UTF16ToString((*[2 << 20]uint16)(unsafe.Pointer(a.AuthorityName))[:])
	return s, nil
}

func (c *Client) KeyInfo() (sessionKeySize uint32, sigAlg uint32, sigAlgName string, encAlg uint32, encAlgName string, err error) {
	var ki _SecPkgContext_KeyInfo
	ret := sspi.QueryContextAttributes(c.ctx.Handle, _SECPKG_ATTR_KEY_INFO, (*byte)(unsafe.Pointer(&ki)))
	if ret != sspi.SEC_E_OK {
		return 0, 0, "", 0, "", ret
	}
	defer sspi.FreeContextBuffer((*byte)(unsafe.Pointer(ki.SignatureAlgorithmName)))
	defer sspi.FreeContextBuffer((*byte)(unsafe.Pointer(ki.EncryptAlgorithmName)))
	saname := syscall.UTF16ToString((*[2 << 20]uint16)(unsafe.Pointer(ki.SignatureAlgorithmName))[:])
	eaname := syscall.UTF16ToString((*[2 << 20]uint16)(unsafe.Pointer(ki.EncryptAlgorithmName))[:])
	return ki.KeySize, ki.SignatureAlgorithm, saname, ki.EncryptAlgorithm, eaname, nil
}

// Sizes queries the context for the sizes used in per-message functions.
// It returns the maximum token size used in authentication exchanges, the
// maximum signature size, the preferred integral size of messages, the
// size of any security trailer, and any error.
func (c *Client) Sizes() (uint32, uint32, uint32, uint32, error) {
	return c.ctx.Sizes()
}
