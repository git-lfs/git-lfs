// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

// Package schannel provides access to the Secure Channel SSP Package.
//
package schannel

import (
	"unsafe"

	"github.com/alexbrainman/sspi"
)

// TODO: add documentation

// PackageInfo contains Secure Channel SSP package description.
var PackageInfo *sspi.PackageInfo

func init() {
	var err error
	PackageInfo, err = sspi.QueryPackageInfo(sspi.UNISP_NAME)
	if err != nil {
		panic("failed to fetch Schannel package info: " + err.Error())
	}
}

func acquireCredentials(creduse uint32) (*sspi.Credentials, error) {
	sc := &__SCHANNEL_CRED{
		Version: __SCHANNEL_CRED_VERSION,
		// TODO: allow for Creds / CredCount
		// TODO: allow for RootStore
		// TODO: allow for EnabledProtocols
		// TODO: allow for MinimumCipherStrength / MaximumCipherStrength
	}
	c, err := sspi.AcquireCredentials(sspi.UNISP_NAME, creduse, (*byte)(unsafe.Pointer(sc)))
	if err != nil {
		return nil, err
	}
	return c, nil
}

func AcquireClientCredentials() (*sspi.Credentials, error) {
	return acquireCredentials(sspi.SECPKG_CRED_OUTBOUND)
}
