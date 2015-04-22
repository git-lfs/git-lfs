package lfs

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"testing"
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
	return "Basic " + base64.URLEncoding.EncodeToString([]byte(token))
}
