package lfsapi

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test that cache satisfies Fill() without looking at creds
func TestCredsCacheFillFromCache(t *testing.T) {
	creds := newFakeCreds()
	cache := withCredentialCache(creds).(*credentialCacher)
	cache.creds["http//lfs.test//foo/bar"] = Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
		"username": "u",
		"password": "p",
	}

	filled, err := cache.Fill(Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
	})
	assert.Nil(t, err)
	require.NotNil(t, filled)
	assert.Equal(t, "u", filled["username"])
	assert.Equal(t, "p", filled["password"])

	assert.Equal(t, 1, len(cache.creds))
	cached, ok := cache.creds["http//lfs.test//foo/bar"]
	assert.True(t, ok)
	assert.Equal(t, "u", cached["username"])
	assert.Equal(t, "p", cached["password"])
}

// test that cache caches Fill() value from creds
func TestCredsCacheFillFromValidHelperFill(t *testing.T) {
	creds := newFakeCreds()
	cache := withCredentialCache(creds).(*credentialCacher)

	creds.list = append(creds.list, Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
		"username": "u",
		"password": "p",
	})

	assert.Equal(t, 0, len(cache.creds))

	filled, err := cache.Fill(Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
	})
	assert.Nil(t, err)
	require.NotNil(t, filled)
	assert.Equal(t, "u", filled["username"])
	assert.Equal(t, "p", filled["password"])

	assert.Equal(t, 1, len(cache.creds))
	cached, ok := cache.creds["http//lfs.test//foo/bar"]
	assert.True(t, ok)
	assert.Equal(t, "u", cached["username"])
	assert.Equal(t, "p", cached["password"])

	creds.list = make([]Creds, 0)
	filled2, err := cache.Fill(Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
	})
	assert.Nil(t, err)
	require.NotNil(t, filled2)
	assert.Equal(t, "u", filled2["username"])
	assert.Equal(t, "p", filled2["password"])
}

// test that cache ignores Fill() value from creds with missing username+password
func TestCredsCacheFillFromInvalidHelperFill(t *testing.T) {
	creds := newFakeCreds()
	cache := withCredentialCache(creds).(*credentialCacher)

	creds.list = append(creds.list, Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
		"username": "no-password",
	})

	assert.Equal(t, 0, len(cache.creds))

	filled, err := cache.Fill(Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
		"username": "u",
		"password": "p",
	})
	assert.Nil(t, err)
	require.NotNil(t, filled)
	assert.Equal(t, "no-password", filled["username"])
	assert.Equal(t, "", filled["password"])

	assert.Equal(t, 0, len(cache.creds))
}

// test that cache ignores Fill() value from creds with error
func TestCredsCacheFillFromErroringHelperFill(t *testing.T) {
	creds := newFakeCreds()
	cache := withCredentialCache(&erroringCreds{creds}).(*credentialCacher)

	creds.list = append(creds.list, Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
		"username": "u",
		"password": "p",
	})

	assert.Equal(t, 0, len(cache.creds))

	filled, err := cache.Fill(Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
	})
	assert.NotNil(t, err)
	require.NotNil(t, filled)
	assert.Equal(t, "u", filled["username"])
	assert.Equal(t, "p", filled["password"])

	assert.Equal(t, 0, len(cache.creds))
}

func TestCredsCacheRejectWithoutError(t *testing.T) {
	creds := newFakeCreds()
	cache := withCredentialCache(creds).(*credentialCacher)

	cache.creds["http//lfs.test//foo/bar"] = Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
		"username": "u",
		"password": "p",
	}

	err := cache.Reject(Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(cache.creds))
}

func TestCredsCacheRejectWithError(t *testing.T) {
	creds := newFakeCreds()
	cache := withCredentialCache(&erroringCreds{creds}).(*credentialCacher)

	cache.creds["http//lfs.test//foo/bar"] = Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
		"username": "u",
		"password": "p",
	}

	err := cache.Reject(Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
	})
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(cache.creds))
}

func TestCredsCacheApproveWithoutError(t *testing.T) {
	creds := newFakeCreds()
	cache := withCredentialCache(creds).(*credentialCacher)

	assert.Equal(t, 0, len(cache.creds))

	err := cache.Approve(Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
		"username": "U",
		"password": "P",
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(cache.creds))
	cached, ok := cache.creds["http//lfs.test//foo/bar"]
	assert.True(t, ok)
	assert.Equal(t, "U", cached["username"])
	assert.Equal(t, "P", cached["password"])
}

func TestCredsCacheApproveWithError(t *testing.T) {
	creds := newFakeCreds()
	cache := withCredentialCache(&erroringCreds{creds}).(*credentialCacher)

	assert.Equal(t, 0, len(cache.creds))

	err := cache.Approve(Creds{
		"protocol": "http",
		"host":     "lfs.test",
		"path":     "foo/bar",
		"username": "u",
		"password": "p",
	})
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(cache.creds))
}

func newFakeCreds() *fakeCreds {
	return &fakeCreds{list: make([]Creds, 0)}
}

type erroringCreds struct {
	helper CredentialHelper
}

func (e *erroringCreds) Fill(creds Creds) (Creds, error) {
	c, _ := e.helper.Fill(creds)
	return c, errors.New("fill error")
}

func (e *erroringCreds) Reject(creds Creds) error {
	e.helper.Reject(creds)
	return errors.New("reject error")
}

func (e *erroringCreds) Approve(creds Creds) error {
	e.helper.Approve(creds)
	return errors.New("approve error")
}

type fakeCreds struct {
	list []Creds
}

func credsMatch(c1, c2 Creds) bool {
	return c1["protocol"] == c2["protocol"] &&
		c1["host"] == c2["host"] &&
		c1["path"] == c2["path"]
}

func (f *fakeCreds) Fill(creds Creds) (Creds, error) {
	for _, saved := range f.list {
		if credsMatch(creds, saved) {
			return saved, nil
		}
	}
	return creds, nil
}

func (f *fakeCreds) Reject(creds Creds) error {
	return nil
}

func (f *fakeCreds) Approve(creds Creds) error {
	return nil
}
