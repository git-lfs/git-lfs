package lfs

import (
	"encoding/base64"
	"net/http"
	"testing"
)

func TestGetCredentialsForAPI(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "https://lfs-server.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := getCredsForAPI(req)
	if err != nil {
		t.Fatal(err)
	}

	if value := creds["username"]; value != "lfs-server.com" {
		t.Errorf("bad username: %s", value)
	}

	if value := creds["password"]; value != "monkey" {
		t.Errorf("bad password: %s", value)
	}

	expected := "Basic " + base64.URLEncoding.EncodeToString([]byte("lfs-server.com:monkey"))
	if value := req.Header.Get("Authorization"); value != expected {
		t.Errorf("Bad Authorization. Expected '%s', got '%s'", expected, value)
	}
}

func TestGetCredentialsForAPIWithExistingAuthorization(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "http://lfs-server.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Test monkey")

	creds, err := getCredsForAPI(req)
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

func TestGetCredentialsForAPIWithSchemeMismatch(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "http://lfs-server.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := getCredsForAPI(req)
	if err != nil {
		t.Fatal(err)
	}

	if creds == nil {
		t.Fatalf("no credentials returned")
	}

	if v := creds["protocol"]; v != "http" {
		t.Errorf("Invalid protocol: %q", v)
	}

	if v := creds["host"]; v != "lfs-server.com" {
		t.Errorf("Invalid host: %q", v)
	}

	if v := creds["path"]; v != "foo" {
		t.Errorf("Invalid path: %q", v)
	}

	expected := "Basic " + base64.URLEncoding.EncodeToString([]byte("lfs-server.com:monkey"))
	if value := req.Header.Get("Authorization"); value != expected {
		t.Errorf("Bad Authorization. Expected '%s', got '%s'", expected, value)
	}
}

func TestGetCredentialsForAPIWithHostMismatch(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "https://lfs-server2.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := getCredsForAPI(req)
	if err != nil {
		t.Fatal(err)
	}

	if v := creds["protocol"]; v != "https" {
		t.Errorf("Invalid protocol: %q", v)
	}

	if v := creds["host"]; v != "lfs-server2.com" {
		t.Errorf("Invalid host: %q", v)
	}

	if v := creds["path"]; v != "foo" {
		t.Errorf("Invalid path: %q", v)
	}

	expected := "Basic " + base64.URLEncoding.EncodeToString([]byte("lfs-server2.com:monkey"))
	if value := req.Header.Get("Authorization"); value != expected {
		t.Errorf("Bad Authorization. Expected '%s', got '%s'", expected, value)
	}
}

func TestGetCredentialsForAPIWithPortMismatch(t *testing.T) {
	Config.SetConfig("lfs.url", "https://lfs-server.com")
	req, err := http.NewRequest("GET", "https://lfs-server.com:8080/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := getCredsForAPI(req)
	if err != nil {
		t.Fatal(err)
	}

	if v := creds["protocol"]; v != "https" {
		t.Errorf("Invalid protocol: %q", v)
	}

	if v := creds["host"]; v != "lfs-server.com:8080" {
		t.Errorf("Invalid host: %q", v)
	}

	if v := creds["path"]; v != "foo" {
		t.Errorf("Invalid path: %q", v)
	}

	expected := "Basic " + base64.URLEncoding.EncodeToString([]byte("lfs-server.com:8080:monkey"))
	if value := req.Header.Get("Authorization"); value != expected {
		t.Errorf("Bad Authorization. Expected '%s', got '%s'", expected, value)
	}
}

func TestGetCredentialsForAPIWithRfc1738UsernameAndPassword(t *testing.T) {
	Config.SetConfig("lfs.url", "https://testuser:testpass@lfs-server.com")
	req, err := http.NewRequest("GET", "https://lfs-server.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := getCredsForAPI(req)
	if err != nil {
		t.Fatal(err)
	}

	if creds != nil {
		t.Errorf("unexpected creds: %v", creds)
	}

	expected := "Basic " + base64.URLEncoding.EncodeToString([]byte("testuser:testpass"))
	if value := req.Header.Get("Authorization"); value != expected {
		t.Errorf("Bad Authorization. Expected '%s', got '%s'", expected, value)
	}
}

func init() {
	execCreds = func(input Creds, subCommand string) (Creds, error) {
		output := make(Creds)
		for key, value := range input {
			output[key] = value
		}
		output["username"] = input["host"]
		output["password"] = "monkey"
		return output, nil
	}
}
