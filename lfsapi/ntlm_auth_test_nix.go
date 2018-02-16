// +build !windows

package lfsapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNtlmClientSession(t *testing.T) {
	cli, err := NewClient(nil)
	require.Nil(t, err)

	creds := &ntmlCredentials{domain: "MOOSEDOMAIN", username: "canadian", password: "MooseAntlersYeah"}
	session1, err := cli.ntlmClientSession(creds)
	assert.Nil(t, err)
	assert.NotNil(t, session1)

	// The second call should ignore creds and give the session we just created.
	badCreds := &ntmlCredentials{domain: "MOOSEDOMAIN", username: "badusername", password: "MooseAntlersYeah"}
	session2, err := cli.ntlmClientSession(badCreds)
	assert.Nil(t, err)
	assert.NotNil(t, session2)
	assert.EqualValues(t, session1, session2)
}

func TestNtlmClientSessionBadCreds(t *testing.T) {
	cli, err := NewClient(nil)
	require.Nil(t, err)

	// Single-Sign-On is not supported on *nix
	_, err = cli.ntlmClientSession(nil)
	assert.NotNil(t, err)
}
