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

func testBasicAdapterExists(t *testing.T) {
	cfg := config.New()
	m := ConfigureManifest(NewManifest(), cfg)

	assert := assert.New(t)

	dls := m.GetDownloadAdapterNames()
	if assert.NotNil(dls) {
		assert.Equal([]string{"basic"}, dls)
	}
	uls := m.GetUploadAdapterNames()
	if assert.NotNil(uls) {
		assert.Equal([]string{"basic"}, uls)
	}

	da := m.NewDownloadAdapter("basic")
	if assert.NotNil(da) {
		assert.Equal("basic", da.Name())
		assert.Equal(Download, da.Direction())
	}

	ua := m.NewUploadAdapter("basic")
	if assert.NotNil(ua) {
		assert.Equal("basic", ua.Name())
		assert.Equal(Upload, ua.Direction())
	}
}

func testAdapterRegAndOverride(t *testing.T) {
	cfg := config.New()
	m := ConfigureManifest(NewManifest(), cfg)
	assert := assert.New(t)

	assert.Nil(m.NewDownloadAdapter("test"))
	assert.Nil(m.NewUploadAdapter("test"))

	m.RegisterNewTransferAdapterFunc("test", Upload, newTestAdapter)
	assert.Nil(m.NewDownloadAdapter("test"))
	assert.NotNil(m.NewUploadAdapter("test"))

	m.RegisterNewTransferAdapterFunc("test", Download, newTestAdapter)
	da := m.NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("test", da.Name())
		assert.Equal(Download, da.Direction())
	}

	ua := m.NewUploadAdapter("test")
	if assert.NotNil(ua) {
		assert.Equal("test", ua.Name())
		assert.Equal(Upload, ua.Direction())
	}

	// Test override
	m.RegisterNewTransferAdapterFunc("test", Upload, newRenamedTestAdapter)
	ua = m.NewUploadAdapter("test")
	if assert.NotNil(ua) {
		assert.Equal("RENAMED", ua.Name())
		assert.Equal(Upload, ua.Direction())
	}

	da = m.NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("test", da.Name())
		assert.Equal(Download, da.Direction())
	}

	m.RegisterNewTransferAdapterFunc("test", Download, newRenamedTestAdapter)
	da = m.NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("RENAMED", da.Name())
		assert.Equal(Download, da.Direction())
	}
}

func testAdapterRegButBasicOnly(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{"lfs.basictransfersonly": "yes"},
	})
	m := ConfigureManifest(NewManifest(), cfg)

	assert := assert.New(t)

	m.RegisterNewTransferAdapterFunc("test", Upload, newTestAdapter)
	m.RegisterNewTransferAdapterFunc("test", Download, newTestAdapter)
	// Will still be created if we ask for them
	assert.NotNil(m.NewUploadAdapter("test"))
	assert.NotNil(m.NewDownloadAdapter("test"))

	// But list will exclude
	ld := m.GetDownloadAdapterNames()
	assert.Equal([]string{BasicAdapterName}, ld)
	lu := m.GetUploadAdapterNames()
	assert.Equal([]string{BasicAdapterName}, lu)
}
