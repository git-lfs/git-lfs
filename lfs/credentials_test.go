package lfs

import (
	"encoding/base64"
	"net/http"
	"testing"
)

func TestGetCredentials(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "https://lfs-server.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := getCreds(req)
	if err != nil {
		t.Fatal(err)
	}

	if value := creds["username"]; value != "lfs-server.com" {
		t.Errorf("bad username: %s", value)
	}

	if value := creds["password"]; value != "monkey" {
		t.Errorf("bad username: %s", value)
	}

	expected := "Basic " + base64.URLEncoding.EncodeToString([]byte("lfs-server.com:monkey"))
	if value := req.Header.Get("Authorization"); value != expected {
		t.Errorf("Bad Authorization. Expected '%s', got '%s'", expected, value)
	}
}

func TestGetCredentialsWithExistingAuthorization(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "http://lfs-server.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Test monkey")

	creds, err := getCreds(req)
	if err != nil {
		t.Fatal(err)
	}

	if creds != nil {
		t.Errorf("Unexpected creds: %v", creds)
	}

	if actual := req.Header.Get("Authorization"); actual != "Test monkey" {
		t.Errorf("Unexpected Authorization header: %s", actual)
	}
}

func TestGetCredentialsWithSchemeMismatch(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "http://lfs-server.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := getCreds(req)
	if err != nil {
		t.Fatal(err)
	}

	if creds != nil {
		t.Errorf("Unexpected creds: %v", creds)
	}

	if actual := req.Header.Get("Authorization"); actual != "" {
		t.Errorf("Unexpected Authorization header: %s", actual)
	}
}

func TestGetCredentialsWithHostMismatch(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "https://lfs-server2.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := getCreds(req)
	if err != nil {
		t.Fatal(err)
	}

	if creds != nil {
		t.Errorf("Unexpected creds: %v", creds)
	}

	if actual := req.Header.Get("Authorization"); actual != "" {
		t.Errorf("Unexpected Authorization header: %s", actual)
	}
}

func TestGetCredentialsWithPortMismatch(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "https://lfs-server:8080.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := getCreds(req)
	if err != nil {
		t.Fatal(err)
	}

	if creds != nil {
		t.Errorf("Unexpected creds: %v", creds)
	}

	if actual := req.Header.Get("Authorization"); actual != "" {
		t.Errorf("Unexpected Authorization header: %s", actual)
	}
}

func init() {
	execCreds = func(input Creds, subCommand string) (credentialFetcher, error) {
		return &testCredentialFetcher{input}, nil
	}
}

type testCredentialFetcher struct {
	Creds Creds
}

func (c *testCredentialFetcher) Credentials() Creds {
	c.Creds["username"] = c.Creds["host"]
	c.Creds["password"] = "monkey"
	return c.Creds
}
