package lfsapi

import (
	"testing"
)

func TestNewEndpointFromCloneURLWithConfig(t *testing.T) {
	expected := "https://foo/bar.git/info/lfs"
	tests := []string{
		"https://foo/bar",
		"https://foo/bar/",
		"https://foo/bar.git",
		"https://foo/bar.git/",
	}

	finder := NewEndpointFinder(nil)
	for _, actual := range tests {
		e := finder.NewEndpointFromCloneURL(actual)
		if e.Url != expected {
			t.Errorf("%s returned bad endpoint url %s", actual, e.Url)
		}
	}
}
