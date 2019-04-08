package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCanonicalization(t *testing.T) {
	vals := map[string][]string{
		"user.name":                   []string{"Pat Doe"},
		"branch.MixedCase.pushremote": []string{"Somewhere"},
		"http.https://example.com/BIG-TEXT.git.extraheader": []string{"X-Foo: Bar"},
	}

	fetcher := GitFetcher{vals: vals}
	assert.Equal(t, []string{"Somewhere"}, fetcher.GetAll("bRanch.MixedCase.pushRemote"))
	assert.Equal(t, []string{"Somewhere"}, fetcher.GetAll("branch.MixedCase.pushremote"))
	assert.Equal(t, []string(nil), fetcher.GetAll("branch.mixedcase.pushremote"))
	assert.Equal(t, []string{"Pat Doe"}, fetcher.GetAll("user.name"))
	assert.Equal(t, []string{"Pat Doe"}, fetcher.GetAll("User.Name"))
	assert.Equal(t, []string{"X-Foo: Bar"}, fetcher.GetAll("http.https://example.com/BIG-TEXT.git.extraheader"))
	assert.Equal(t, []string{"X-Foo: Bar"}, fetcher.GetAll("http.https://example.com/BIG-TEXT.git.extraHeader"))
	assert.Equal(t, []string(nil), fetcher.GetAll("http.https://example.com/big-text.git.extraHeader"))
}
