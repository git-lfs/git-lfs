package hawserclient

import (
	"github.com/bmizerany/assert"
	"github.com/hawser/git-hawser/hawser"
	"os"
	"path/filepath"
	"testing"
)

func TestObjectUrl(t *testing.T) {
	oid := "oid"
	tests := map[string]string{
		"http://example.com":      "http://example.com/objects/oid",
		"http://example.com/":     "http://example.com/objects/oid",
		"http://example.com/foo":  "http://example.com/foo/objects/oid",
		"http://example.com/foo/": "http://example.com/foo/objects/oid",
	}

	config := hawser.Config
	for endpoint, expected := range tests {
		config.SetConfig("hawser.url", endpoint)
		assert.Equal(t, expected, ObjectUrl(oid).String())
	}
}

var TmpPath string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	TmpPath = filepath.Join(wd, "..", "tmp", "hawserclient")
	os.MkdirAll(TmpPath, 0755)

	execCreds = func(input Creds, subCommand string) (credentialFetcher, error) {
		return &testCredentialFetcher{input}, nil
	}
}

type testCredentialFetcher struct {
	Creds Creds
}

func (c *testCredentialFetcher) Credentials() Creds {
	return c.Creds
}
