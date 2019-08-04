// +build darwin

package tools

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckCloneFileSupported(t *testing.T) {
	as := assert.New(t)

	// Do
	ok, err := CheckCloneFileSupported(os.TempDir())

	// Verify
	t.Logf("ok = %v, err = %v", ok, err) // Just logging for 1st element

	if !checkCloneFileSupported() {
		as.EqualError(err, "unsupported OS version. >= 10.12.x Sierra required")
	}
}

func TestCloneFile(t *testing.T) {
	as := assert.New(t)

	// Do
	ok, err := CloneFile(nil, nil)

	// Verify always no error and not ok
	as.NoError(err)
	as.False(ok)
}

func TestCloneFileByPath(t *testing.T) {
	if !cloneFileSupported {
		t.Skip("clone not supported on this platform")
	}

	src := path.Join(os.TempDir(), "src")
	t.Logf("src = %s", src)

	dst := path.Join(os.TempDir(), "dst")
	t.Logf("dst = %s", dst)

	as := assert.New(t)

	// Precondition
	err := ioutil.WriteFile(src, []byte("TEST"), 0666)
	as.NoError(err)

	// Do
	ok, err := CloneFileByPath(dst, src)
	if err != nil {
		if cloneFileError, ok := err.(*CloneFileError); ok && cloneFileError.Unsupported {
			t.Log(err)
			t.Skip("tmp file is not support clonefile in this os installation.")
		}
		t.Error(err)
	}

	// Verify
	as.NoError(err)
	as.True(ok)

	dstContents, err := ioutil.ReadFile(dst)
	as.NoError(err)
	as.Equal("TEST", string(dstContents))
}
