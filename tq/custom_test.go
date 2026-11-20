package tq

import (
	"testing"

	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/stretchr/testify/assert"
)

func TestCustomTransferBasicConfig(t *testing.T) {
	path := "/path/to/binary"
	cli := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.customtransfer.testsimple.path": path,
	}))
	defer cli.Close()

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

func TestCustomTransferPathConfigMatchesCompleteKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "exact",
			key:      "lfs.customtransfer.testcustom.path",
			expected: true,
		},
		{
			name: "access suffix",
			key:  "lfs.customtransfer.testcustom.path.access",
		},
		{
			name: "access suffix with extra prefix",
			key:  "extra.lfs.customtransfer.testcustom.path.access",
		},
		{
			name: "remote prefix",
			key:  "remote.lfs.customtransfer.testcustom.path",
		},
		{
			name: "remote prefix with extra suffix",
			key:  "remote.lfs.customtransfer.testcustom.path.extra",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cli := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
				test.key: "/path/to/binary",
			}))
			defer cli.Close()

			m := NewManifest(nil, cli, "", "").Upgrade()
			assert.Equal(t, test.expected, m.isCustomAdapter("testcustom", "download"))
			assert.Equal(t, test.expected, m.isCustomAdapter("testcustom", "upload"))
		})
	}
}

func TestCustomTransferDownloadConfig(t *testing.T) {
	path := "/path/to/binary"
	args := "-c 1 --whatever"
	cli := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.customtransfer.testdownload.path":       path,
		"lfs.customtransfer.testdownload.args":       args,
		"lfs.customtransfer.testdownload.concurrent": "false",
		"lfs.customtransfer.testdownload.direction":  "download",
	}))
	defer cli.Close()

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
	cli := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.customtransfer.testupload.path":       path,
		"lfs.customtransfer.testupload.args":       args,
		"lfs.customtransfer.testupload.concurrent": "false",
		"lfs.customtransfer.testupload.direction":  "upload",
	}))
	defer cli.Close()

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
	cli := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.customtransfer.testboth.path":       path,
		"lfs.customtransfer.testboth.args":       args,
		"lfs.customtransfer.testboth.concurrent": "yes",
		"lfs.customtransfer.testboth.direction":  "both",
	}))
	defer cli.Close()

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
