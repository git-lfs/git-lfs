// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package sspi_test

import (
	"testing"

	"github.com/alexbrainman/sspi"
)

func TestQueryPackageInfo(t *testing.T) {
	pkgnames := []string{
		sspi.NTLMSP_NAME,
		sspi.MICROSOFT_KERBEROS_NAME,
		sspi.NEGOSSP_NAME,
		sspi.UNISP_NAME,
	}
	for _, name := range pkgnames {
		pi, err := sspi.QueryPackageInfo(name)
		if err != nil {
			t.Error(err)
			continue
		}
		if pi.Name != name {
			t.Errorf("unexpected package name %q returned for %q package: package info is %#v", pi.Name, name, pi)
			continue
		}
	}
}
