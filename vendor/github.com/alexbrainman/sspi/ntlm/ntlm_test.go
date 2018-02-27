// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package ntlm_test

import (
	"flag"
	"os/user"
	"runtime"
	"testing"
	"time"

	"github.com/alexbrainman/sspi"
	"github.com/alexbrainman/sspi/ntlm"
)

var (
	testDomain   = flag.String("domain", "", "domain parameter for TestAcquireUserCredentials")
	testUsername = flag.String("username", "", "username parameter for TestAcquireUserCredentials")
	testPassword = flag.String("password", "", "password parameter for TestAcquireUserCredentials")
)

func TestPackageInfo(t *testing.T) {
	if ntlm.PackageInfo.Name != "NTLM" {
		t.Fatalf(`invalid NTLM package name of %q, "NTLM" is expected.`, ntlm.PackageInfo.Name)
	}
}

func testContextExpiry(t *testing.T, name string, c interface {
	Expiry() time.Time
}) {
	validFor := c.Expiry().Sub(time.Now())
	if validFor < time.Hour {
		t.Errorf("%v exipries in %v, more then 1 hour expected", name, validFor)
	}
	if validFor > 10*24*time.Hour {
		t.Errorf("%v exipries in %v, less then 10 days expected", name, validFor)
	}
}

func testNTLM(t *testing.T, clientCred *sspi.Credentials) {
	serverCred, err := ntlm.AcquireServerCredentials()
	if err != nil {
		t.Fatal(err)
	}
	defer serverCred.Release()

	client, token1, err := ntlm.NewClientContext(clientCred)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Release()

	testContextExpiry(t, "clent security context", client)

	server, token2, err := ntlm.NewServerContext(serverCred, token1)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Release()

	testContextExpiry(t, "server security context", server)

	token3, err := client.Update(token2)
	if err != nil {
		t.Fatal(err)
	}

	err = server.Update(token3)
	if err != nil {
		t.Fatal(err)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err = server.ImpersonateUser()
	if err != nil {
		t.Fatal(err)
	}
	defer server.RevertToSelf()

	_, err = user.Current()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNTLM(t *testing.T) {
	cred, err := ntlm.AcquireCurrentUserCredentials()
	if err != nil {
		t.Fatal(err)
	}
	defer cred.Release()

	testNTLM(t, cred)
}

func TestAcquireUserCredentials(t *testing.T) {
	if len(*testDomain) == 0 {
		t.Skip("Skipping due to empty \"domain\" parameter")
	}
	if len(*testUsername) == 0 {
		t.Skip("Skipping due to empty \"username\" parameter")
	}
	if len(*testPassword) == 0 {
		t.Skip("Skipping due to empty \"password\" parameter")
	}
	cred, err := ntlm.AcquireUserCredentials(*testDomain, *testUsername, *testPassword)
	if err != nil {
		t.Fatal(err)
	}
	defer cred.Release()

	testNTLM(t, cred)
}
