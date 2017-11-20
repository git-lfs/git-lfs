package lfsapi

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
