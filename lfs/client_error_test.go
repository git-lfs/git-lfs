package lfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestSuccessStatus(t *testing.T) {
	for _, status := range []int{200, 201, 202} {
		res := &http.Response{StatusCode: status}
		if err := handleResponse(res); err != nil {
			t.Errorf("Unexpected error for HTTP %d: %s", status, err.Error())
		}
	}
}

func TestErrorStatusWithCustomMessage(t *testing.T) {
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

		wErr := handleResponse(res)
		if wErr == nil {
			t.Errorf("No error from HTTP %d", status)
			continue
		}

		expected := fmt.Sprintf("custom error for %d", status)
		if actual := wErr.Error(); actual != expected {
			t.Errorf("Expected for HTTP %d:\n%s\nACTUAL:\n%s", status, expected, actual)
			continue
		}

		if wErr.Panic == (panicMsg != "panic") {
			t.Errorf("Error for HTTP %d should %s", status, panicMsg)
			continue
		}
	}
}

func TestErrorStatusWithDefaultMessage(t *testing.T) {
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
		429: {defaultErrors[400] + " from HTTP 429", "not panic"},
		500: {defaultErrors[500], "panic"},
		501: {defaultErrors[500] + " from HTTP 501", "not panic"},
		503: {defaultErrors[500] + " from HTTP 503", "panic"},
		504: {defaultErrors[500] + " from HTTP 504", "panic"},
		509: {defaultErrors[500] + " from HTTP 509", "not panic"},
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

		wErr := handleResponse(res)
		if wErr == nil {
			t.Errorf("No error from HTTP %d", status)
			continue
		}

		expected := fmt.Sprintf(results[0], rawurl)

		if actual := wErr.Error(); actual != expected {
			t.Errorf("Expected for HTTP %d:\n%s\nACTUAL:\n%s", status, expected, actual)
			continue
		}

		if wErr.Panic == (results[1] != "panic") {
			t.Errorf("Error for HTTP %d should %s", status, results[1])
			continue
		}
	}
}
