package transfer

import (
	"testing"

	"github.com/github/git-lfs/config"

	"github.com/stretchr/testify/assert"
)

type testAdapter struct {
	name string
	dir  Direction
}

func (a *testAdapter) Name() string {
	return a.name
}
func (a *testAdapter) Direction() Direction {
	return a.dir
}
func (a *testAdapter) Begin(maxConcurrency int, cb TransferProgressCallback, completion chan TransferResult) error {
	return nil
}
func (a *testAdapter) Add(t *Transfer) {
}
func (a *testAdapter) End() {
}
func (a *testAdapter) ClearTempStorage() error {
	return nil
}
func newTestAdapter(name string, dir Direction) TransferAdapter {
	return &testAdapter{name, dir}
}
func newRenamedTestAdapter(name string, dir Direction) TransferAdapter {
	return &testAdapter{"RENAMED", dir}
}
func resetAdapters() {
	uploadAdapterFuncs = make(map[string]NewTransferAdapterFunc)
	downloadAdapterFuncs = make(map[string]NewTransferAdapterFunc)
}

func testBasicAdapterExists(t *testing.T) {
	assert := assert.New(t)

	dls := GetDownloadAdapterNames()
	if assert.NotNil(dls) {
		assert.Equal([]string{"basic"}, dls)
	}
	uls := GetUploadAdapterNames()
	if assert.NotNil(uls) {
		assert.Equal([]string{"basic"}, uls)
	}
	da := NewDownloadAdapter("basic")
	if assert.NotNil(da) {
		assert.Equal("basic", da.Name())
		assert.Equal(Download, da.Direction())
	}
	ua := NewUploadAdapter("basic")
	if assert.NotNil(ua) {
		assert.Equal("basic", ua.Name())
		assert.Equal(Upload, ua.Direction())
	}
}

func testAdapterRegAndOverride(t *testing.T) {
	assert := assert.New(t)

	assert.Nil(NewDownloadAdapter("test"))
	assert.Nil(NewUploadAdapter("test"))

	RegisterNewTransferAdapterFunc("test", Upload, newTestAdapter)
	assert.Nil(NewDownloadAdapter("test"))
	assert.NotNil(NewUploadAdapter("test"))

	RegisterNewTransferAdapterFunc("test", Download, newTestAdapter)
	da := NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("test", da.Name())
		assert.Equal(Download, da.Direction())
	}
	ua := NewUploadAdapter("test")
	if assert.NotNil(ua) {
		assert.Equal("test", ua.Name())
		assert.Equal(Upload, ua.Direction())
	}

	// Test override
	RegisterNewTransferAdapterFunc("test", Upload, newRenamedTestAdapter)
	ua = NewUploadAdapter("test")
	if assert.NotNil(ua) {
		assert.Equal("RENAMED", ua.Name())
		assert.Equal(Upload, ua.Direction())
	}
	da = NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("test", da.Name())
		assert.Equal(Download, da.Direction())
	}
	RegisterNewTransferAdapterFunc("test", Download, newRenamedTestAdapter)
	da = NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("RENAMED", da.Name())
		assert.Equal(Download, da.Direction())
	}

}

func testAdapterRegButBasicOnly(t *testing.T) {
	assert := assert.New(t)

	config.Config.SetConfig("lfs.basictransfersonly", "yes")
	RegisterNewTransferAdapterFunc("test", Upload, newTestAdapter)
	RegisterNewTransferAdapterFunc("test", Download, newTestAdapter)
	// Will still be created if we ask for them
	assert.NotNil(NewUploadAdapter("test"))
	assert.NotNil(NewDownloadAdapter("test"))

	// But list will exclude
	ld := GetDownloadAdapterNames()
	assert.Equal([]string{BasicAdapterName}, ld)
	lu := GetUploadAdapterNames()
	assert.Equal([]string{BasicAdapterName}, lu)
}
