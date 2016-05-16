package lfs

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/github/git-lfs/auth"
)

var (
	TestCredentialsFunc auth.CredentialFunc
	origCredentialsFunc auth.CredentialFunc
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
	TestCredentialsFunc = func(input auth.Creds, subCommand string) (auth.Creds, error) {
		output := make(auth.Creds)
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
	origCredentialsFunc = auth.SetCredentialsFunc(TestCredentialsFunc)
}

// Put the original credentials func back
func RestoreCredentialsFunc() {
	auth.SetCredentialsFunc(origCredentialsFunc)
}
