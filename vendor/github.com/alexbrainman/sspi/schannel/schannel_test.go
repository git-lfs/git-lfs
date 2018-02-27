// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package schannel_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"testing"

	"github.com/alexbrainman/sspi/schannel"
)

func TestPackageInfo(t *testing.T) {
	want := "Microsoft Unified Security Protocol Provider"
	if schannel.PackageInfo.Name != want {
		t.Fatalf(`invalid Schannel package name of %q, %q is expected.`, schannel.PackageInfo.Name, want)
	}
}

func TestSchannel(t *testing.T) {
	cred, err := schannel.AcquireClientCredentials()
	if err != nil {
		t.Fatal(err)
	}
	defer cred.Release()

	conn, err := net.Dial("tcp", "microsoft.com:https")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := schannel.NewClientContext(cred, conn)
	err = client.Handshake("microsoft.com")
	if err != nil {
		t.Fatal(err)
	}
	protoName, major, minor, err := client.ProtocolInfo()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("protocol info: %s %d.%d", protoName, major, minor)
	userName, err := client.UserName()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("user name: %q", userName)
	authorityName, err := client.AuthorityName()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("authority name: %q", authorityName)
	sessionKeySize, sigAlg, sigAlgName, encAlg, encAlgName, err := client.KeyInfo()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("key info: session_key_size=%d signature_alg=%q(%d) encryption_alg=%q(%d)", sessionKeySize, sigAlgName, sigAlg, encAlgName, encAlg)
	// TODO: add some code to verify if negotiated connection is suitable (ciper and so on)
	_, err = fmt.Fprintf(client, "GET / HTTP/1.1\r\nHost: foo\r\n\r\n")
	if err != nil {
		t.Fatal(err)
	}
	data, err := ioutil.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("web page: %q", data)
	err = client.Shutdown()
	if err != nil {
		t.Fatal(err)
	}
}
