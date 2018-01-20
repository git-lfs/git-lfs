package tq

import (
	"testing"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func (a *testAdapter) Begin(cfg AdapterConfig, cb ProgressCallback) error {
	return nil
}

func (a *testAdapter) Add(ts ...*Transfer) (retries <-chan TransferResult) {
	return nil
}

func (a *testAdapter) End() {
}

func (a *testAdapter) ClearTempStorage() error {
	return nil
}

func newTestAdapter(name string, dir Direction) Adapter {
	return &testAdapter{name, dir}
}

func newRenamedTestAdapter(name string, dir Direction) Adapter {
	return &testAdapter{"RENAMED", dir}
}

func testBasicAdapterExists(t *testing.T) {
	m := NewManifest(nil, nil, "", "")

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
	m := NewManifest(nil, nil, "", "")
	assert := assert.New(t)

	assert.Nil(m.NewDownloadAdapter("test"))
	assert.Nil(m.NewUploadAdapter("test"))

	m.RegisterNewAdapterFunc("test", Upload, newTestAdapter)
	assert.Nil(m.NewDownloadAdapter("test"))
	assert.NotNil(m.NewUploadAdapter("test"))

	m.RegisterNewAdapterFunc("test", Download, newTestAdapter)
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
	m.RegisterNewAdapterFunc("test", Upload, newRenamedTestAdapter)
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

	m.RegisterNewAdapterFunc("test", Download, newRenamedTestAdapter)
	da = m.NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("RENAMED", da.Name())
		assert.Equal(Download, da.Direction())
	}
}

func testAdapterRegButBasicOnly(t *testing.T) {
	cli, err := lfsapi.NewClient(lfsapi.NewContext(nil, nil, map[string]string{
		"lfs.basictransfersonly": "yes",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")

	assert := assert.New(t)

	m.RegisterNewAdapterFunc("test", Upload, newTestAdapter)
	m.RegisterNewAdapterFunc("test", Download, newTestAdapter)
	// Will still be created if we ask for them
	assert.NotNil(m.NewUploadAdapter("test"))
	assert.NotNil(m.NewDownloadAdapter("test"))

	// But list will exclude
	ld := m.GetDownloadAdapterNames()
	assert.Equal([]string{BasicAdapterName}, ld)
	lu := m.GetUploadAdapterNames()
	assert.Equal([]string{BasicAdapterName}, lu)
}
