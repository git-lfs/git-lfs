package transfer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/github/git-lfs/config"
)

var (
	savedDownloadAdapterFuncs map[string]NewTransferAdapterFunc
	savedUploadAdapterFuncs   map[string]NewTransferAdapterFunc
)

func copyFuncMap(to, from map[string]NewTransferAdapterFunc) {
	for k, v := range from {
		to[k] = v
	}
}

func saveTransferSetupState() {
	funcMutex.Lock()
	defer funcMutex.Unlock()

	savedDownloadAdapterFuncs = make(map[string]NewTransferAdapterFunc)
	copyFuncMap(savedDownloadAdapterFuncs, downloadAdapterFuncs)

	savedUploadAdapterFuncs = make(map[string]NewTransferAdapterFunc)
	copyFuncMap(savedUploadAdapterFuncs, uploadAdapterFuncs)
}

func restoreTransferSetupState() {
	funcMutex.Lock()
	defer funcMutex.Unlock()

	downloadAdapterFuncs = make(map[string]NewTransferAdapterFunc)
	copyFuncMap(downloadAdapterFuncs, savedDownloadAdapterFuncs)

	uploadAdapterFuncs = make(map[string]NewTransferAdapterFunc)
	copyFuncMap(uploadAdapterFuncs, savedUploadAdapterFuncs)

}

func TestCustomTransferBasicConfig(t *testing.T) {
	saveTransferSetupState()
	defer func() {
		config.Config.ResetConfig()
		restoreTransferSetupState()
	}()

	path := "/path/to/binary"
	config.Config.SetConfig("lfs.customtransfer.testsimple.path", path)

	ConfigureCustomAdapters()

	u := NewUploadAdapter("testsimple")
	assert.NotNil(t, u, "Upload adapter should be present")
	cu, _ := u.(*customAdapter)
	assert.NotNil(t, cu, "Upload adapter should be customAdapter")
	assert.Equal(t, cu.path, path, "Path should be correct")
	assert.Equal(t, cu.args, "", "args should be blank")
	assert.Equal(t, cu.concurrent, true, "concurrent should be defaulted")

	d := NewDownloadAdapter("testsimple")
	assert.NotNil(t, d, "Download adapter should be present")
	cd, _ := u.(*customAdapter)
	assert.NotNil(t, cd, "Download adapter should be customAdapter")
	assert.Equal(t, cd.path, path, "Path should be correct")
	assert.Equal(t, cd.args, "", "args should be blank")
	assert.Equal(t, cd.concurrent, true, "concurrent should be defaulted")
}

func TestCustomTransferDownloadConfig(t *testing.T) {
	saveTransferSetupState()
	defer func() {
		config.Config.ResetConfig()
		restoreTransferSetupState()
	}()

	path := "/path/to/binary"
	args := "-c 1 --whatever"
	config.Config.SetConfig("lfs.customtransfer.testdownload.path", path)
	config.Config.SetConfig("lfs.customtransfer.testdownload.args", args)
	config.Config.SetConfig("lfs.customtransfer.testdownload.concurrent", "false")
	config.Config.SetConfig("lfs.customtransfer.testdownload.direction", "download")

	ConfigureCustomAdapters()

	u := NewUploadAdapter("testdownload")
	assert.NotNil(t, u, "Upload adapter should always be created")
	cu, _ := u.(*customAdapter)
	assert.Nil(t, cu, "Upload adapter should NOT be custom (default to basic)")

	d := NewDownloadAdapter("testdownload")
	assert.NotNil(t, d, "Download adapter should be present")
	cd, _ := d.(*customAdapter)
	assert.NotNil(t, cd, "Download adapter should be customAdapter")
	assert.Equal(t, cd.path, path, "Path should be correct")
	assert.Equal(t, cd.args, args, "args should be correct")
	assert.Equal(t, cd.concurrent, false, "concurrent should be set")
}

func TestCustomTransferUploadConfig(t *testing.T) {
	saveTransferSetupState()
	defer func() {
		config.Config.ResetConfig()
		restoreTransferSetupState()
	}()

	path := "/path/to/binary"
	args := "-c 1 --whatever"
	config.Config.SetConfig("lfs.customtransfer.testupload.path", path)
	config.Config.SetConfig("lfs.customtransfer.testupload.args", args)
	config.Config.SetConfig("lfs.customtransfer.testupload.concurrent", "false")
	config.Config.SetConfig("lfs.customtransfer.testupload.direction", "upload")

	ConfigureCustomAdapters()

	d := NewDownloadAdapter("testupload")
	assert.NotNil(t, d, "Download adapter should always be created")
	cd, _ := d.(*customAdapter)
	assert.Nil(t, cd, "Download adapter should NOT be custom (default to basic)")

	u := NewUploadAdapter("testupload")
	assert.NotNil(t, u, "Upload adapter should be present")
	cu, _ := u.(*customAdapter)
	assert.NotNil(t, cu, "Upload adapter should be customAdapter")
	assert.Equal(t, cu.path, path, "Path should be correct")
	assert.Equal(t, cu.args, args, "args should be correct")
	assert.Equal(t, cu.concurrent, false, "concurrent should be set")
}

func TestCustomTransferBothConfig(t *testing.T) {
	saveTransferSetupState()
	defer func() {
		config.Config.ResetConfig()
		restoreTransferSetupState()
	}()

	path := "/path/to/binary"
	args := "-c 1 --whatever --yeah"
	config.Config.SetConfig("lfs.customtransfer.testboth.path", path)
	config.Config.SetConfig("lfs.customtransfer.testboth.args", args)
	config.Config.SetConfig("lfs.customtransfer.testboth.concurrent", "yes")
	config.Config.SetConfig("lfs.customtransfer.testboth.direction", "both")

	ConfigureCustomAdapters()

	d := NewDownloadAdapter("testboth")
	assert.NotNil(t, d, "Download adapter should be present")
	cd, _ := d.(*customAdapter)
	assert.NotNil(t, cd, "Download adapter should be customAdapter")
	assert.Equal(t, cd.path, path, "Path should be correct")
	assert.Equal(t, cd.args, args, "args should be correct")
	assert.Equal(t, cd.concurrent, true, "concurrent should be set")

	u := NewUploadAdapter("testboth")
	assert.NotNil(t, u, "Upload adapter should be present")
	cu, _ := u.(*customAdapter)
	assert.NotNil(t, cu, "Upload adapter should be customAdapter")
	assert.Equal(t, cu.path, path, "Path should be correct")
	assert.Equal(t, cu.args, args, "args should be correct")
	assert.Equal(t, cu.concurrent, true, "concurrent should be set")
}
