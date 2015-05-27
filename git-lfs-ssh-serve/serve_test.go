package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"github.com/github/git-lfs/lfs"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmizerany/assert"
)

// Here we use a real client SSH context to talk to the real Serve() function
// However a Pipe is used to connect the two, no real SSH at this point
func TestServe(t *testing.T) {

	testcontentsz := int64(634)

	config := NewConfig()
	config.BasePath = filepath.Join(os.TempDir(), "git-lfs-serve-test")
	os.MkdirAll(config.BasePath, 0755)
	repopath := "test/repo"

	testcontent := make([]byte, testcontentsz)
	// put something interesting in it so we can detect it at each end
	testcontent[0] = '2'
	testcontent[1] = 'Z'
	testcontent[2] = '>'
	testcontent[3] = 'Q'
	testcontent[testcontentsz-1] = '#'
	testcontent[testcontentsz-2] = 'y'
	testcontent[testcontentsz-3] = 'L'
	testcontent[testcontentsz-4] = 'A'

	// Defer cleanup
	defer os.RemoveAll(config.BasePath)

	hasher := sha256.New()
	inbuf := bytes.NewReader(testcontent)
	io.Copy(hasher, inbuf)
	testoid := hex.EncodeToString(hasher.Sum(nil))

	cli, srv := net.Pipe()
	var outerr bytes.Buffer

	// 'Serve' is the real server function, usually connected to stdin/stdout but to pipe for test
	go Serve(srv, srv, &outerr, config, repopath)
	defer cli.Close()

	ctx := lfs.NewManualSSHApiContext(cli, cli)

	// Upload chunk (no callback used, that's tested in client tests)
	cb := func(int64, int64, int) error { return nil }
	rdr := bytes.NewReader(testcontent)

	wrerr := ctx.Upload(testoid, int64(len(testcontent)), rdr, cb)
	if wrerr != nil { // assert.Equal doesn't seem to work with WrappedError
		t.Errorf("Error in upload: %v", wrerr)
	}
	assert.Equal(t, 0, rdr.Len()) // server should have read all bytes
	uploadDestPath, _ := mediaPath(testoid, config)
	s, err := os.Stat(uploadDestPath)
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(len(testcontent)), s.Size())

	// Prove that it fails safely when trying to upload duplicate content
	rdr = bytes.NewReader(testcontent)
	wrerr = ctx.Upload(testoid, int64(len(testcontent)), rdr, cb)
	if wrerr == nil { // assert.Equal doesn't seem to work with WrappedError
		t.Errorf("Upload should have safely errored")
	}

	// Now try to download same data
	var dlbuf bytes.Buffer
	dlrdr, sz, wrerr := ctx.Download(testoid)
	if wrerr != nil {
		t.Errorf("Error in download: %v", wrerr)
	}
	assert.Equal(t, sz, testcontentsz)
	_, err = io.CopyN(&dlbuf, dlrdr, sz)
	if err != nil {
		t.Errorf("Error copying from download", err)
	}

	downloadedbytes := dlbuf.Bytes()
	assert.Equal(t, testcontent, downloadedbytes)

	// Now try to download a non-existent oid to make sure it fails safely
	dlbuf.Reset()
	dlrdr, sz, wrerr = ctx.Download("99999999999999999999999999999999999")
	if wrerr == nil {
		t.Errorf("Should have failed safely when downloading missing")
	}

	ctx.Close()

}
