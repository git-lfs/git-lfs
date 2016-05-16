package lfs

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/github/git-lfs/credentials"
)

var (
	TestCredentialsFunc credentials.CredentialFunc
	origCredentialsFunc credentials.CredentialFunc
)

func tempdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "git-lfs-test")
	if err != nil {
		t.Fatalf("Error getting temp dir: %s", err)
	}
	return dir
}

func expectedAuth(t *testing.T, server *httptest.Server) string {
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	token := fmt.Sprintf("%s:%s", u.Host, "monkey")
	return "Basic " + strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte(token)))
}

func init() {
	TestCredentialsFunc = func(input credentials.Creds, subCommand string) (credentials.Creds, error) {
		output := make(credentials.Creds)
		for key, value := range input {
			output[key] = value
		}
		if _, ok := output["username"]; !ok {
			output["username"] = input["host"]
		}
		output["password"] = "monkey"
		return output, nil
	}
}

// Override the credentials func for testing
func SetupTestCredentialsFunc() {
	origCredentialsFunc = credentials.SetCredentialsFunc(TestCredentialsFunc)
}

// Put the original credentials func back
func RestoreCredentialsFunc() {
	credentials.SetCredentialsFunc(origCredentialsFunc)
}
