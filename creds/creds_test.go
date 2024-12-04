package creds

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertCredsLinesMatch(t *testing.T, expected []string, buf *bytes.Buffer) {
	expected = append(expected, "")
	actual := strings.SplitAfter(buf.String(), "\n")

	slices.Sort(expected)
	slices.Sort(actual)

	assert.Equal(t, expected, actual)
}

func TestCredsBufferFormat(t *testing.T) {
	creds := make(Creds)

	expected := []string{"capability[]=authtype\n", "capability[]=state\n"}

	buf, err := creds.buffer(true)
	assert.NoError(t, err)
	assertCredsLinesMatch(t, expected, buf)

	creds["protocol"] = []string{"https"}
	creds["host"] = []string{"example.com"}

	expectedPrefix := strings.Join(expected, "")
	expected = append(expected, "protocol=https\n", "host=example.com\n")

	buf, err = creds.buffer(true)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(buf.String(), expectedPrefix))
	assertCredsLinesMatch(t, expected, buf)

	creds["wwwauth[]"] = []string{"Basic realm=test", "Negotiate"}

	expected = append(expected, "wwwauth[]=Basic realm=test\n")
	expected = append(expected, "wwwauth[]=Negotiate\n")

	buf, err = creds.buffer(true)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(buf.String(), expectedPrefix))
	assertCredsLinesMatch(t, expected, buf)
}

func TestCredsBufferProtect(t *testing.T) {
	creds := make(Creds)

	// Always disallow LF characters
	creds["protocol"] = []string{"https"}
	creds["host"] = []string{"one.example.com\nhost=two.example.com"}

	buf, err := creds.buffer(false)
	assert.Error(t, err)
	assert.Nil(t, buf)

	buf, err = creds.buffer(true)
	assert.Error(t, err)
	assert.Nil(t, buf)

	// Disallow CR characters unless protocol protection disabled
	creds["host"] = []string{"one.example.com\rhost=two.example.com"}

	expected := []string{
		"capability[]=authtype\n",
		"capability[]=state\n",
		"protocol=https\n",
		"host=one.example.com\rhost=two.example.com\n",
	}

	buf, err = creds.buffer(false)
	assert.NoError(t, err)
	assertCredsLinesMatch(t, expected, buf)

	buf, err = creds.buffer(true)
	assert.Error(t, err)
	assert.Nil(t, buf)

	// Always disallow null bytes
	creds["host"] = []string{"one.example.com\x00host=two.example.com"}

	buf, err = creds.buffer(false)
	assert.Error(t, err)
	assert.Nil(t, buf)

	buf, err = creds.buffer(true)
	assert.Error(t, err)
	assert.Nil(t, buf)
}

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
	cache := NewCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": []string{"https"}, "host": []string{"example.com"}}

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

	credsWithPass := Creds{
		"protocol": []string{"https"},
		"host":     []string{"example.com"},
		"username": []string{"foo"},
		"password": []string{"bar"},
	}
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
	cache := NewCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": []string{"https"}, "host": []string{"example.com"}}

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
	cache := NewCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": []string{"https"}, "host": []string{"example.com"}}

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
	cache := NewCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": []string{"https"}, "host": []string{"example.com"}}

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
	cache := NewCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": []string{"https"}, "host": []string{"example.com"}}

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
	cache := NewCredentialCacher()
	helper1 := newTestCredHelper()
	helper2 := newTestCredHelper()
	helpers := NewCredentialHelpers([]CredentialHelper{cache, helper1, helper2})
	creds := Creds{"protocol": []string{"https"}, "host": []string{"example.com"}}

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
