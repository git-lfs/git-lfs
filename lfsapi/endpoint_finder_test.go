package lfsapi

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/git-lfs/git-lfs/v3/creds"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/stretchr/testify/assert"
)

func TestEndpointDefaultsToOrigin(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.lfsurl": "abc",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "abc", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestEndpointOverridesOrigin(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url":              "abc",
		"remote.origin.lfsurl": "def",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "abc", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestEndpointNoOverrideDefaultRemote(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.lfsurl": "abc",
		"remote.other.lfsurl":  "def",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "abc", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestEndpointUseAlternateRemote(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.lfsurl": "abc",
		"remote.other.lfsurl":  "def",
	}))

	e := finder.Endpoint("download", "other")
	assert.Equal(t, "def", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "https://example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestBareEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "https://example.com/foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestEndpointSeparateClonePushUrl(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url":     "https://example.com/foo/bar.git",
		"remote.origin.pushurl": "https://readwrite.com/foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)

	e = finder.Endpoint("upload", "")
	assert.Equal(t, "https://readwrite.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestEndpointOverriddenSeparateClonePushLfsUrl(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url":        "https://example.com/foo/bar.git",
		"remote.origin.pushurl":    "https://readwrite.com/foo/bar.git",
		"remote.origin.lfsurl":     "https://examplelfs.com/foo/bar",
		"remote.origin.lfspushurl": "https://readwritelfs.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://examplelfs.com/foo/bar", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)

	e = finder.Endpoint("upload", "")
	assert.Equal(t, "https://readwritelfs.com/foo/bar", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestEndpointGlobalSeparateLfsPush(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url":     "https://readonly.com/foo/bar",
		"lfs.pushurl": "https://write.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://readonly.com/foo/bar", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)

	e = finder.Endpoint("upload", "")
	assert.Equal(t, "https://write.com/foo/bar", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestSSHEndpointOverridden(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url":    "git@example.com:foo/bar",
		"remote.origin.lfsurl": "lfs",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestSSHEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "ssh://git@example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/foo/bar", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "ssh", e.SSHMetadata.Scheme)
}

func TestSSHCustomPortEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "ssh://git@example.com:9000/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/foo/bar", e.SSHMetadata.Path)
	assert.Equal(t, "9000", e.SSHMetadata.Port)
	assert.Equal(t, "ssh", e.SSHMetadata.Scheme)
}

func TestGitSSHEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "git+ssh://git@example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/foo/bar", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "git+ssh", e.SSHMetadata.Scheme)
}

func TestGitSSHCustomPortEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "git+ssh://git@example.com:9000/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/foo/bar", e.SSHMetadata.Path)
	assert.Equal(t, "9000", e.SSHMetadata.Port)
	assert.Equal(t, "git+ssh", e.SSHMetadata.Scheme)
}

func TestSSHGitEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "ssh+git://git@example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/foo/bar", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "ssh+git", e.SSHMetadata.Scheme)
}

func TestSSHGitCustomPortEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "ssh+git://git@example.com:9000/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/foo/bar", e.SSHMetadata.Path)
	assert.Equal(t, "9000", e.SSHMetadata.Port)
	assert.Equal(t, "ssh+git", e.SSHMetadata.Scheme)
}

func TestBareSSHEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "git@example.com:foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "foo/bar.git", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestBareSSSHEndpointWithCustomPortInBrackets(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "[git@example.com:2222]:foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "foo/bar.git", e.SSHMetadata.Path)
	assert.Equal(t, "2222", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

/**** DEBUG chrisd -
func TestBareSSSHEndpointWithCustomPortInBracketsAfterUser(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "git@[example.com:2222]:foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "foo/bar.git", e.SSHMetadata.Path)
	assert.Equal(t, "2222", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}
*/

func TestSSHEndpointFromGlobalLfsUrl(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url": "git@example.com:foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git", e.Url)
	assert.Equal(t, "git@example.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "foo/bar.git", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestHTTPEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "http://example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestBareHTTPEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "http://example.com/foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestGitEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "git://example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestGitEndpointAddsLfsSuffixWithCustomProtocol(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "git://example.com/foo/bar",
		"lfs.gitprotocol":   "http",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestBareGitEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": "git://example.com/foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)
	assert.Equal(t, "", e.SSHMetadata.Port)
	assert.Equal(t, "", e.SSHMetadata.Scheme)
}

func TestLocalPathEndpointAddsDotGitForWorkingRepo(t *testing.T) {
	// Windows will add a drive letter to the paths below since we
	// canonicalize them.
	if runtime.GOOS == "windows" {
		return
	}

	path, err := ioutil.TempDir("", "lfsRepo")
	assert.Nil(t, err)
	path = path + "/local/path"
	err = os.MkdirAll(path+"/.git", 0755)
	assert.Nil(t, err)

	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": path,
	}))
	e := finder.Endpoint("download", "")
	assert.Equal(t, "file://"+path+"/.git", e.Url)

	os.RemoveAll(path)
}

func TestLocalPathEndpointPreservesDotGitForWorkingRepo(t *testing.T) {
	// Windows will add a drive letter to the paths below since we
	// canonicalize them.
	if runtime.GOOS == "windows" {
		return
	}

	path, err := ioutil.TempDir("", "lfsRepo")
	assert.Nil(t, err)
	path = path + "/local/path/.git"
	err = os.MkdirAll(path, 0755)
	assert.Nil(t, err)

	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": path,
	}))
	e := finder.Endpoint("download", "")
	assert.Equal(t, "file://"+path, e.Url)

	os.RemoveAll(path)
}

func TestLocalPathEndpointPreservesNoDotGitForBareRepo(t *testing.T) {
	// Windows will add a drive letter to the paths below since we
	// canonicalize them.
	if runtime.GOOS == "windows" {
		return
	}

	path, err := ioutil.TempDir("", "lfsRepo")
	assert.Nil(t, err)
	path = path + "/local/path"
	err = os.MkdirAll(path, 0755)
	assert.Nil(t, err)

	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": path,
	}))
	e := finder.Endpoint("download", "")
	assert.Equal(t, "file://"+path, e.Url)

	os.RemoveAll(path)
}

func TestLocalPathEndpointRemovesDotGitForBareRepo(t *testing.T) {
	// Windows will add a drive letter to the paths below since we
	// canonicalize them.
	if runtime.GOOS == "windows" {
		return
	}

	path, err := ioutil.TempDir("", "lfsRepo")
	assert.Nil(t, err)
	path = path + "/local/path"
	err = os.MkdirAll(path, 0755)
	assert.Nil(t, err)

	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.origin.url": path + "/.git",
	}))
	e := finder.Endpoint("download", "")
	assert.Equal(t, "file://"+path, e.Url)

	os.RemoveAll(path)
}

func TestAccessConfig(t *testing.T) {
	type accessTest struct {
		AccessMode    string
		PrivateAccess bool
	}

	tests := map[string]accessTest{
		"":            {"none", false},
		"basic":       {"basic", true},
		"BASIC":       {"basic", true},
		"private":     {"basic", true},
		"PRIVATE":     {"basic", true},
		"invalidauth": {"invalidauth", true},
	}

	for value, expected := range tests {
		finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
			"lfs.url":                        "http://example.com",
			"lfs.http://example.com.access":  value,
			"lfs.https://example.com.access": "bad",
		}))

		dl := finder.Endpoint("upload", "")
		ul := finder.Endpoint("download", "")

		if access := finder.AccessFor(dl.Url); access.Mode() != creds.AccessMode(expected.AccessMode) {
			t.Errorf("Expected creds.AccessMode() with value %q to be %v, got %v", value, expected.AccessMode, access)
		}
		if access := finder.AccessFor(ul.Url); access.Mode() != creds.AccessMode(expected.AccessMode) {
			t.Errorf("Expected creds.AccessMode() with value %q to be %v, got %v", value, expected.AccessMode, access)
		}
	}

	// Test again but with separate push url
	for value, expected := range tests {
		finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
			"lfs.url":                           "http://example.com",
			"lfs.pushurl":                       "http://examplepush.com",
			"lfs.http://example.com.access":     value,
			"lfs.http://examplepush.com.access": value,
			"lfs.https://example.com.access":    "bad",
		}))

		dl := finder.Endpoint("upload", "")
		ul := finder.Endpoint("download", "")

		if access := finder.AccessFor(dl.Url); access.Mode() != creds.AccessMode(expected.AccessMode) {
			t.Errorf("Expected creds.AccessMode() with value %q to be %v, got %v", value, expected.AccessMode, access)
		}
		if access := finder.AccessFor(ul.Url); access.Mode() != creds.AccessMode(expected.AccessMode) {
			t.Errorf("Expected creds.AccessMode() with value %q to be %v, got %v", value, expected.AccessMode, access)
		}
	}
}

func TestAccessAbsentConfig(t *testing.T) {
	finder := NewEndpointFinder(nil)

	downloadAccess := finder.AccessFor(finder.Endpoint("download", "").Url)
	assert.Equal(t, creds.NoneAccess, downloadAccess.Mode())

	uploadAccess := finder.AccessFor(finder.Endpoint("upload", "").Url)
	assert.Equal(t, creds.NoneAccess, uploadAccess.Mode())
}

func TestSetAccess(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{}))
	url := "http://example.com"
	access := finder.AccessFor(url)

	assert.Equal(t, creds.NoneAccess, access.Mode())
	assert.Equal(t, url, access.URL())

	finder.SetAccess(access.Upgrade(creds.NegotiateAccess))

	newAccess := finder.AccessFor(url)
	assert.Equal(t, creds.NegotiateAccess, newAccess.Mode())
	assert.Equal(t, url, newAccess.URL())
}

func TestChangeAccess(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.http://example.com.access": "basic",
	}))

	url := "http://example.com"
	access := finder.AccessFor(url)
	assert.Equal(t, creds.BasicAccess, access.Mode())
	assert.Equal(t, url, access.URL())

	finder.SetAccess(access.Upgrade(creds.NegotiateAccess))

	newAccess := finder.AccessFor(url)
	assert.Equal(t, creds.NegotiateAccess, newAccess.Mode())
	assert.Equal(t, url, newAccess.URL())
}

func TestDeleteAccessWithNone(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.http://example.com.access": "basic",
	}))

	url := "http://example.com"

	access := finder.AccessFor(url)
	assert.Equal(t, creds.BasicAccess, access.Mode())
	assert.Equal(t, url, access.URL())

	finder.SetAccess(access.Upgrade(creds.NoneAccess))

	newAccess := finder.AccessFor(url)
	assert.Equal(t, creds.NoneAccess, newAccess.Mode())
	assert.Equal(t, url, newAccess.URL())
}

func TestDeleteAccessWithEmptyString(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.http://example.com.access": "basic",
	}))

	url := "http://example.com"

	access := finder.AccessFor(url)
	assert.Equal(t, creds.BasicAccess, access.Mode())
	assert.Equal(t, url, access.URL())

	finder.SetAccess(access.Upgrade(creds.AccessMode("")))

	newAccess := finder.AccessFor(url)
	assert.Equal(t, creds.NoneAccess, newAccess.Mode())
	assert.Equal(t, url, newAccess.URL())
}

type EndpointParsingTestCase struct {
	Given    string
	Expected lfshttp.Endpoint
}

func (c *EndpointParsingTestCase) Assert(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"url.https://github.com/.insteadof": "gh:",
	}))
	actual := finder.NewEndpoint("upload", c.Given)
	assert.Equal(t, c.Expected, actual, "lfsapi: expected endpoint for %q to be %#v (was %#v)", c.Given, c.Expected, actual)
}

func TestEndpointParsing(t *testing.T) {
	// Note that many of these tests will produce silly or completely broken
	// values for the Url, and that's okay: they work nevertheless.
	for desc, c := range map[string]EndpointParsingTestCase{
		"simple bare ssh": {
			"git@github.com:git-lfs/git-lfs.git",
			lfshttp.Endpoint{
				Url: "https://github.com/git-lfs/git-lfs.git",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "git@github.com",
					Path:        "git-lfs/git-lfs.git",
					Port:        "",
					Scheme:      "",
				},
				Operation: "",
			},
		},
		"port bare ssh": {
			"[git@lfshttp.github.com:443]:git-lfs/git-lfs.git",
			lfshttp.Endpoint{
				Url: "https://lfshttp.github.com/git-lfs/git-lfs.git",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "git@lfshttp.github.com",
					Path:        "git-lfs/git-lfs.git",
					Port:        "443",
					Scheme:      "",
				},
				Operation: "",
			},
		},
//// DEBUG chrisd - add git@[ test
		"no user bare ssh": {
			"github.com:git-lfs/git-lfs.git",
			lfshttp.Endpoint{
				Url: "https://github.com/git-lfs/git-lfs.git",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "github.com",
					Path:        "git-lfs/git-lfs.git",
					Port:        "",
					Scheme:      "",
				},
				Operation: "",
			},
		},
		"bare word bare ssh": {
			"github:git-lfs/git-lfs.git",
			lfshttp.Endpoint{
				Url: "https://github/git-lfs/git-lfs.git",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "github",
					Path:        "git-lfs/git-lfs.git",
					Port:        "",
					Scheme:      "",
				},
				Operation: "",
			},
		},
		"insteadof alias": {
			"gh:git-lfs/git-lfs.git",
			lfshttp.Endpoint{
				Url: "https://github.com/git-lfs/git-lfs.git",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "",
					Path:        "",
					Port:        "",
					Scheme:      "",
				},
				Operation: "",
			},
		},
		"remote helper": {
			"remote::git-lfs/git-lfs.git",
			lfshttp.Endpoint{
				Url: "remote::git-lfs/git-lfs.git",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "",
					Path:        "",
					Port:        "",
					Scheme:      "",
				},
				Operation: "",
			},
		},
	} {
		t.Run(desc, c.Assert)
	}
}

type InsteadOfTestCase struct {
	Given     string
	Operation string
	Expected  lfshttp.Endpoint
}

func (c *InsteadOfTestCase) Assert(t *testing.T) {
	finder := NewEndpointFinder(lfshttp.NewContext(nil, nil, map[string]string{
		"remote.test.url":                      c.Given,
		"url.https://example.com/.insteadof":   "ex:",
		"url.ssh://example.com/.pushinsteadof": "ex:",
		"url.ssh://example.com/.insteadof":     "exp:",
	}))
	actual := finder.Endpoint(c.Operation, "test")
	assert.Equal(t, c.Expected, actual, "lfsapi: expected endpoint for %q to be %#v (was %#v)", c.Given, c.Expected, actual)
}

func TestInsteadOf(t *testing.T) {
	// Note that many of these tests will produce silly or completely broken
	// values for the Url, and that's okay: they work nevertheless.
	for desc, c := range map[string]InsteadOfTestCase{
		"insteadof alias (download)": {
			"ex:git-lfs/git-lfs.git",
			"download",
			lfshttp.Endpoint{
				Url: "https://example.com/git-lfs/git-lfs.git/info/lfs",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "",
					Path:        "",
					Port:        "",
					Scheme:      "",
				},
				Operation: "download",
			},
		},
		"pushinsteadof alias (upload)": {
			"ex:git-lfs/git-lfs.git",
			"upload",
			lfshttp.Endpoint{
				Url: "https://example.com/git-lfs/git-lfs.git/info/lfs",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "example.com",
					Path:        "/git-lfs/git-lfs.git",
					Port:        "",
					Scheme:      "ssh",
				},
				Operation: "upload",
			},
		},
		"exp alias (download)": {
			"exp:git-lfs/git-lfs.git",
			"download",
			lfshttp.Endpoint{
				Url: "https://example.com/git-lfs/git-lfs.git/info/lfs",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "example.com",
					Path:        "/git-lfs/git-lfs.git",
					Port:        "",
					Scheme:      "ssh",
				},
				Operation: "download",
			},
		},
		"exp alias (upload)": {
			"exp:git-lfs/git-lfs.git",
			"upload",
			lfshttp.Endpoint{
				Url: "https://example.com/git-lfs/git-lfs.git/info/lfs",
				SSHMetadata: ssh.SSHMetadata{
					UserAndHost: "example.com",
					Path:        "/git-lfs/git-lfs.git",
					Port:        "",
					Scheme:      "ssh",
				},
				Operation: "upload",
			},
		},
	} {
		t.Run(desc, c.Assert)
	}
}

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
		e := finder.NewEndpointFromCloneURL("upload", actual)
		if e.Url != expected {
			t.Errorf("%s returned bad endpoint url %s", actual, e.Url)
		}
	}
}
