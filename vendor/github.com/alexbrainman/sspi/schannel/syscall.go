// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package schannel

import (
	"syscall"
)

// TODO: maybe put all these into a separate package, like sspi/schannel/winapi or similar

const (
	__SCHANNEL_CRED_VERSION = 4

	_SP_PROT_PCT1_SERVER      = 0x00000001
	_SP_PROT_PCT1_CLIENT      = 0x00000002
	_SP_PROT_PCT1             = _SP_PROT_PCT1_SERVER | _SP_PROT_PCT1_CLIENT
	_SP_PROT_SSL2_SERVER      = 0x00000004
	_SP_PROT_SSL2_CLIENT      = 0x00000008
	_SP_PROT_SSL2             = _SP_PROT_SSL2_SERVER | _SP_PROT_SSL2_CLIENT
	_SP_PROT_SSL3_SERVER      = 0x00000010
	_SP_PROT_SSL3_CLIENT      = 0x00000020
	_SP_PROT_SSL3             = _SP_PROT_SSL3_SERVER | _SP_PROT_SSL3_CLIENT
	_SP_PROT_TLS1_SERVER      = 0x00000040
	_SP_PROT_TLS1_CLIENT      = 0x00000080
	_SP_PROT_TLS1             = _SP_PROT_TLS1_SERVER | _SP_PROT_TLS1_CLIENT
	_SP_PROT_SSL3TLS1_CLIENTS = _SP_PROT_TLS1_CLIENT | _SP_PROT_SSL3_CLIENT
	_SP_PROT_SSL3TLS1_SERVERS = _SP_PROT_TLS1_SERVER | _SP_PROT_SSL3_SERVER
	_SP_PROT_SSL3TLS1         = _SP_PROT_SSL3 | _SP_PROT_TLS1
)

type __SCHANNEL_CRED struct {
	Version               uint32
	CredCount             uint32
	Creds                 *syscall.CertContext
	RootStore             syscall.Handle // TODO: make sure this field is syscall.Handle
	cMappers              uint32
	aphMappers            uintptr
	SupportedAlgCount     uint32
	SupportedAlgs         *uint32
	EnabledProtocols      uint32
	MinimumCipherStrength uint32
	MaximumCipherStrength uint32
	SessionLifespan       uint32
	Flags                 uint32
	CredFormat            uint32
}

const (
	_SECPKG_ATTR_SIZES            = 0
	_SECPKG_ATTR_NAMES            = 1
	_SECPKG_ATTR_LIFESPAN         = 2
	_SECPKG_ATTR_DCE_INFO         = 3
	_SECPKG_ATTR_STREAM_SIZES     = 4
	_SECPKG_ATTR_KEY_INFO         = 5
	_SECPKG_ATTR_AUTHORITY        = 6
	_SECPKG_ATTR_PROTO_INFO       = 7
	_SECPKG_ATTR_PASSWORD_EXPIRY  = 8
	_SECPKG_ATTR_SESSION_KEY      = 9
	_SECPKG_ATTR_PACKAGE_INFO     = 10
	_SECPKG_ATTR_USER_FLAGS       = 11
	_SECPKG_ATTR_NEGOTIATION_INFO = 12
	_SECPKG_ATTR_NATIVE_NAMES     = 13
	_SECPKG_ATTR_FLAGS            = 14

	_SCHANNEL_RENEGOTIATE = 0
	_SCHANNEL_SHUTDOWN    = 1
	_SCHANNEL_ALERT       = 2
)

type _SecPkgContext_StreamSizes struct {
	Header         uint32
	Trailer        uint32
	MaximumMessage uint32
	Buffers        uint32
	BlockSize      uint32
}

type _SecPkgContext_ProtoInfo struct {
	ProtocolName *uint16
	MajorVersion uint32
	MinorVersion uint32
}

type _SecPkgContext_Names struct {
	UserName *uint16
}

type _SecPkgContext_Authority struct {
	AuthorityName *uint16
}

type _SecPkgContext_KeyInfo struct {
	SignatureAlgorithmName *uint16
	EncryptAlgorithmName   *uint16
	KeySize                uint32
	SignatureAlgorithm     uint32
	EncryptAlgorithm       uint32
}

// TODO: SecPkgContext_ConnectionInfo

// TODO: SECPKG_ATTR_REMOTE_CERT_CONTEXT
// TODO: SECPKG_ATTR_LOCAL_CERT_CONTEXT

// TODO: SecPkgContext_IssuerListInfoEx
