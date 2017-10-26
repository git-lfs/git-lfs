package tq

import (
	"testing"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomTransferBasicConfig(t *testing.T) {
	path := "/path/to/binary"
	cli, err := lfsapi.NewClient(lfsapi.NewContext(nil, nil, map[string]string{
		"lfs.customtransfer.testsimple.path": path,
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	u := m.NewUploadAdapter("testsimple")
	assert.NotNil(t, u, "Upload adapter should be present")
	cu, _ := u.(*customAdapter)
	assert.NotNil(t, cu, "Upload adapter should be customAdapter")
	assert.Equal(t, cu.path, path, "Path should be correct")
	assert.Equal(t, cu.args, "", "args should be blank")
	assert.Equal(t, cu.concurrent, true, "concurrent should be defaulted")

	d := m.NewDownloadAdapter("testsimple")
	assert.NotNil(t, d, "Download adapter should be present")
	cd, _ := u.(*customAdapter)
	assert.NotNil(t, cd, "Download adapter should be customAdapter")
	assert.Equal(t, cd.path, path, "Path should be correct")
	assert.Equal(t, cd.args, "", "args should be blank")
	assert.Equal(t, cd.concurrent, true, "concurrent should be defaulted")
}

func TestCustomTransferDownloadConfig(t *testing.T) {
	path := "/path/to/binary"
	args := "-c 1 --whatever"
	cli, err := lfsapi.NewClient(lfsapi.NewContext(nil, nil, map[string]string{
		"lfs.customtransfer.testdownload.path":       path,
		"lfs.customtransfer.testdownload.args":       args,
		"lfs.customtransfer.testdownload.concurrent": "false",
		"lfs.customtransfer.testdownload.direction":  "download",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	u := m.NewUploadAdapter("testdownload")
	assert.NotNil(t, u, "Upload adapter should always be created")
	cu, _ := u.(*customAdapter)
	assert.Nil(t, cu, "Upload adapter should NOT be custom (default to basic)")

	d := m.NewDownloadAdapter("testdownload")
	assert.NotNil(t, d, "Download adapter should be present")
	cd, _ := d.(*customAdapter)
	assert.NotNil(t, cd, "Download adapter should be customAdapter")
	assert.Equal(t, cd.path, path, "Path should be correct")
	assert.Equal(t, cd.args, args, "args should be correct")
	assert.Equal(t, cd.concurrent, false, "concurrent should be set")
}

func TestCustomTransferUploadConfig(t *testing.T) {
	path := "/path/to/binary"
	args := "-c 1 --whatever"
	cli, err := lfsapi.NewClient(lfsapi.NewContext(nil, nil, map[string]string{
		"lfs.customtransfer.testupload.path":       path,
		"lfs.customtransfer.testupload.args":       args,
		"lfs.customtransfer.testupload.concurrent": "false",
		"lfs.customtransfer.testupload.direction":  "upload",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	d := m.NewDownloadAdapter("testupload")
	assert.NotNil(t, d, "Download adapter should always be created")
	cd, _ := d.(*customAdapter)
	assert.Nil(t, cd, "Download adapter should NOT be custom (default to basic)")

	u := m.NewUploadAdapter("testupload")
	assert.NotNil(t, u, "Upload adapter should be present")
	cu, _ := u.(*customAdapter)
	assert.NotNil(t, cu, "Upload adapter should be customAdapter")
	assert.Equal(t, cu.path, path, "Path should be correct")
	assert.Equal(t, cu.args, args, "args should be correct")
	assert.Equal(t, cu.concurrent, false, "concurrent should be set")
}

func TestCustomTransferBothConfig(t *testing.T) {
	path := "/path/to/binary"
	args := "-c 1 --whatever --yeah"
	cli, err := lfsapi.NewClient(lfsapi.NewContext(nil, nil, map[string]string{
		"lfs.customtransfer.testboth.path":       path,
		"lfs.customtransfer.testboth.args":       args,
		"lfs.customtransfer.testboth.concurrent": "yes",
		"lfs.customtransfer.testboth.direction":  "both",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	d := m.NewDownloadAdapter("testboth")
	assert.NotNil(t, d, "Download adapter should be present")
	cd, _ := d.(*customAdapter)
	assert.NotNil(t, cd, "Download adapter should be customAdapter")
	assert.Equal(t, cd.path, path, "Path should be correct")
	assert.Equal(t, cd.args, args, "args should be correct")
	assert.Equal(t, cd.concurrent, true, "concurrent should be set")

	u := m.NewUploadAdapter("testboth")
	assert.NotNil(t, u, "Upload adapter should be present")
	cu, _ := u.(*customAdapter)
	assert.NotNil(t, cu, "Upload adapter should be customAdapter")
	assert.Equal(t, cu.path, path, "Path should be correct")
	assert.Equal(t, cu.args, args, "args should be correct")
	assert.Equal(t, cu.concurrent, true, "concurrent should be set")
}
