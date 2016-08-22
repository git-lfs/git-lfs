package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errors"
)

func TestSuccessStatus(t *testing.T) {
	cfg := config.New()
	for _, status := range []int{200, 201, 202} {
		res := &http.Response{StatusCode: status}
		if err := handleResponse(cfg, res, nil); err != nil {
			t.Errorf("Unexpected error for HTTP %d: %s", status, err.Error())
		}
	}
}

func TestErrorStatusWithCustomMessage(t *testing.T) {
	cfg := config.New()
	u, err := url.Parse("https://lfs-server.com/objects/oid")
	if err != nil {
		t.Fatal(err)
	}

	statuses := map[int]string{
		400: "not panic",
		401: "not panic",
		403: "not panic",
		404: "not panic",
		405: "not panic",
		406: "not panic",
		429: "not panic",
		500: "panic",
		501: "not panic",
		503: "panic",
		504: "panic",
		507: "not panic",
		509: "not panic",
	}

	for status, panicMsg := range statuses {
		cliErr := &ClientError{
			Message: fmt.Sprintf("custom error for %d", status),
		}

		by, err := json.Marshal(cliErr)
		if err != nil {
			t.Errorf("Error building json for status %d: %s", status, err)
			continue
		}

		res := &http.Response{
			StatusCode: status,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewReader(by)),
			Request:    &http.Request{URL: u},
		}
		res.Header.Set("Content-Type", "application/vnd.git-lfs+json; charset=utf-8")

		err = handleResponse(cfg, res, nil)
		if err == nil {
			t.Errorf("No error from HTTP %d", status)
			continue
		}

		expected := fmt.Sprintf("custom error for %d", status)
		if actual := err.Error(); !strings.HasSuffix(actual, expected) {
			t.Errorf("Expected for HTTP %d:\n%s\nACTUAL:\n%s", status, expected, actual)
			continue
		}

		if errors.IsFatalError(err) == (panicMsg != "panic") {
			t.Errorf("Error for HTTP %d should %s", status, panicMsg)
			continue
		}
	}
}

func TestErrorStatusWithDefaultMessage(t *testing.T) {
	cfg := config.New()
	rawurl := "https://lfs-server.com/objects/oid"
	u, err := url.Parse(rawurl)
	if err != nil {
		t.Fatal(err)
	}

	statuses := map[int][]string{
		400: {defaultErrors[400], "not panic"},
		401: {defaultErrors[401], "not panic"},
		403: {defaultErrors[401], "not panic"},
		404: {defaultErrors[404], "not panic"},
		405: {defaultErrors[400] + " from HTTP 405", "not panic"},
		406: {defaultErrors[400] + " from HTTP 406", "not panic"},
		429: {defaultErrors[429], "not panic"},
		500: {defaultErrors[500], "panic"},
		501: {defaultErrors[500] + " from HTTP 501", "not panic"},
		503: {defaultErrors[500] + " from HTTP 503", "panic"},
		504: {defaultErrors[500] + " from HTTP 504", "panic"},
		507: {defaultErrors[507], "not panic"},
		509: {defaultErrors[509], "not panic"},
	}

	for status, results := range statuses {
		cliErr := &ClientError{
			Message: fmt.Sprintf("custom error for %d", status),
		}

		by, err := json.Marshal(cliErr)
		if err != nil {
			t.Errorf("Error building json for status %d: %s", status, err)
			continue
		}

		res := &http.Response{
			StatusCode: status,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewReader(by)),
			Request:    &http.Request{URL: u},
		}

		// purposely wrong content type so it falls back to default
		res.Header.Set("Content-Type", "application/vnd.git-lfs+json2")

		err = handleResponse(cfg, res, nil)
		if err == nil {
			t.Errorf("No error from HTTP %d", status)
			continue
		}

		expected := fmt.Sprintf(results[0], rawurl)
		if actual := err.Error(); !strings.HasSuffix(actual, expected) {
			t.Errorf("Expected for HTTP %d:\n%s\nACTUAL:\n%s", status, expected, actual)
			continue
		}

		if errors.IsFatalError(err) == (results[1] != "panic") {
			t.Errorf("Error for HTTP %d should %s", status, results[1])
			continue
		}
	}
}
