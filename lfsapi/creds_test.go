package lfsapi

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCredHelper struct {
	fillErr    error
	approveErr error
	rejectErr  error
	fill       []Creds
	approve    []Creds
	reject     []Creds
}

func newTestCredHelper() *testCredHelper {
	return &testCredHelper{
		fill:    make([]Creds, 0),
		approve: make([]Creds, 0),
		reject:  make([]Creds, 0),
	}
}

func (h *testCredHelper) Fill(input Creds) (Creds, error) {
	h.fill = append(h.fill, input)
	return input, h.fillErr
}

func (h *testCredHelper) Approve(creds Creds) error {
	h.approve = append(h.approve, creds)
	return h.approveErr
}

func (h *testCredHelper) Reject(creds Creds) error {
	h.reject = append(h.reject, creds)
	return h.rejectErr
}

func TestCredHelperSetNoErrors(t *testing.T) {
	cache := newCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": "https", "host": "example.com"}

	out, err := helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 1, len(helper1.fill))
	assert.Equal(t, 0, len(helper2.fill))

	// calling Fill() with empty cache
	out, err = helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 2, len(helper1.fill))
	assert.Equal(t, 0, len(helper2.fill))

	credsWithPass := Creds{"protocol": "https", "host": "example.com", "username": "foo", "password": "bar"}
	assert.Nil(t, helpers.Approve(credsWithPass))
	assert.Equal(t, 1, len(helper1.approve))
	assert.Equal(t, 0, len(helper2.approve))

	// calling Approve() again is cached
	assert.Nil(t, helpers.Approve(credsWithPass))
	assert.Equal(t, 1, len(helper1.approve))
	assert.Equal(t, 0, len(helper2.approve))

	// access cache
	for i := 0; i < 3; i++ {
		out, err = helpers.Fill(creds)
		assert.Nil(t, err)
		assert.Equal(t, credsWithPass, out)
		assert.Equal(t, 2, len(helper1.fill))
		assert.Equal(t, 0, len(helper2.fill))
	}

	assert.Nil(t, helpers.Reject(creds))
	assert.Equal(t, 1, len(helper1.reject))
	assert.Equal(t, 0, len(helper2.reject))

	// Reject() is never cached
	assert.Nil(t, helpers.Reject(creds))
	assert.Equal(t, 2, len(helper1.reject))
	assert.Equal(t, 0, len(helper2.reject))

	// calling Fill() with empty cache
	out, err = helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 3, len(helper1.fill))
	assert.Equal(t, 0, len(helper2.fill))
}

func TestCredHelperSetFillError(t *testing.T) {
	cache := newCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": "https", "host": "example.com"}

	helper1.fillErr = errors.New("boom")
	out, err := helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 1, len(helper1.fill))
	assert.Equal(t, 1, len(helper2.fill))

	assert.Nil(t, helpers.Approve(creds))
	assert.Equal(t, 0, len(helper1.approve))
	assert.Equal(t, 1, len(helper2.approve))

	// Fill() with cache
	for i := 0; i < 3; i++ {
		out, err = helpers.Fill(creds)
		assert.Nil(t, err)
		assert.Equal(t, creds, out)
		assert.Equal(t, 1, len(helper1.fill))
		assert.Equal(t, 1, len(helper2.fill))
	}

	assert.Nil(t, helpers.Reject(creds))
	assert.Equal(t, 0, len(helper1.reject))
	assert.Equal(t, 1, len(helper2.reject))

	// Fill() with empty cache
	out, err = helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 1, len(helper1.fill)) // still skipped
	assert.Equal(t, 2, len(helper2.fill))
}

func TestCredHelperSetApproveError(t *testing.T) {
	cache := newCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": "https", "host": "example.com"}

	approveErr := errors.New("boom")
	helper1.approveErr = approveErr
	out, err := helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 1, len(helper1.fill))
	assert.Equal(t, 0, len(helper2.fill))

	assert.Equal(t, approveErr, helpers.Approve(creds))
	assert.Equal(t, 1, len(helper1.approve))
	assert.Equal(t, 0, len(helper2.approve))

	// cache is never set
	out, err = helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 2, len(helper1.fill))
	assert.Equal(t, 0, len(helper2.fill))

	assert.Nil(t, helpers.Reject(creds))
	assert.Equal(t, 1, len(helper1.reject))
	assert.Equal(t, 0, len(helper2.reject))
}

func TestCredHelperSetFillAndApproveError(t *testing.T) {
	cache := newCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": "https", "host": "example.com"}

	credErr := errors.New("boom")
	helper1.fillErr = credErr
	helper2.approveErr = credErr

	out, err := helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 1, len(helper1.fill))
	assert.Equal(t, 1, len(helper2.fill))

	assert.Equal(t, credErr, helpers.Approve(creds))
	assert.Equal(t, 0, len(helper1.approve)) // skipped
	assert.Equal(t, 0, len(helper1.reject))  // skipped
	assert.Equal(t, 1, len(helper2.approve))

	// never approved, so cache is empty
	out, err = helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 1, len(helper1.fill)) // still skipped
	assert.Equal(t, 2, len(helper2.fill))
}

func TestCredHelperSetRejectError(t *testing.T) {
	cache := newCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": "https", "host": "example.com"}

	rejectErr := errors.New("boom")
	helper1.rejectErr = rejectErr
	out, err := helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 1, len(helper1.fill))
	assert.Equal(t, 0, len(helper2.fill))

	assert.Nil(t, helpers.Approve(creds))
	assert.Equal(t, 1, len(helper1.approve))
	assert.Equal(t, 0, len(helper2.approve))

	// Fill() with cache
	out, err = helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 1, len(helper1.fill))
	assert.Equal(t, 0, len(helper2.fill))

	assert.Equal(t, rejectErr, helpers.Reject(creds))
	assert.Equal(t, 1, len(helper1.reject))
	assert.Equal(t, 0, len(helper2.reject))

	// failed Reject() still clears cache
	out, err = helpers.Fill(creds)
	assert.Nil(t, err)
	assert.Equal(t, creds, out)
	assert.Equal(t, 2, len(helper1.fill))
	assert.Equal(t, 0, len(helper2.fill))
}

func TestCredHelperSetAllFillErrors(t *testing.T) {
	cache := newCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": "https", "host": "example.com"}

	helper1.fillErr = errors.New("boom 1")
	helper2.fillErr = errors.New("boom 2")
	out, err := helpers.Fill(creds)
	if assert.NotNil(t, err) {
		assert.Equal(t, "credential fill errors:\nboom 1\nboom 2", err.Error())
	}
	assert.Nil(t, out)
	assert.Equal(t, 1, len(helper1.fill))
	assert.Equal(t, 1, len(helper2.fill))

	err = helpers.Approve(creds)
	if assert.NotNil(t, err) {
		assert.Equal(t, "no valid credential helpers to approve", err.Error())
	}
	assert.Equal(t, 0, len(helper1.approve))
	assert.Equal(t, 0, len(helper2.approve))

	err = helpers.Reject(creds)
	if assert.NotNil(t, err) {
		assert.Equal(t, "no valid credential helpers to reject", err.Error())
	}
	assert.Equal(t, 0, len(helper1.reject))
	assert.Equal(t, 0, len(helper2.reject))
}

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
