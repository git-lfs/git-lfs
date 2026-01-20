package tq

import (
	"testing"

	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
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

func newTestAdapter(name string, dir Direction) Adapter {
	return &testAdapter{name, dir}
}

func newRenamedTestAdapter(name string, dir Direction) Adapter {
	return &testAdapter{"RENAMED", dir}
}

func TestBasicAdapterExists(t *testing.T) {
	m := NewManifest(nil, nil, "", "")

	assert := assert.New(t)

	dls := m.GetDownloadAdapterNames()
	if assert.NotNil(dls) {
		assert.ElementsMatch([]string{"basic", "lfs-standalone-file", "ssh"}, dls)
	}
	uls := m.GetUploadAdapterNames()
	if assert.NotNil(uls) {
		assert.ElementsMatch([]string{"basic", "lfs-standalone-file", "ssh"}, dls)
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

func TestAdapterRegAndOverride(t *testing.T) {
	m := NewManifest(nil, nil, "", "")
	assert := assert.New(t)

	assert.Nil(m.NewAdapter("test", Download))
	assert.Nil(m.NewAdapter("test", Upload))

	da := m.NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("basic", da.Name())
		assert.Equal(Download, da.Direction())
	}

	ua := m.NewUploadAdapter("test")
	if assert.NotNil(ua) {
		assert.Equal("basic", ua.Name())
		assert.Equal(Upload, ua.Direction())
	}

	m.RegisterNewAdapterFunc("test", Upload, newTestAdapter)
	assert.Nil(m.NewAdapter("test", Download))
	assert.NotNil(m.NewAdapter("test", Upload))

	da = m.NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("basic", da.Name())
		assert.Equal(Download, da.Direction())
	}

	ua = m.NewUploadAdapter("test")
	if assert.NotNil(ua) {
		assert.Equal("test", ua.Name())
		assert.Equal(Upload, ua.Direction())
	}

	m.RegisterNewAdapterFunc("test", Download, newTestAdapter)
	assert.NotNil(m.NewAdapter("test", Download))
	assert.NotNil(m.NewAdapter("test", Upload))

	da = m.NewDownloadAdapter("test")
	if assert.NotNil(da) {
		assert.Equal("test", da.Name())
		assert.Equal(Download, da.Direction())
	}

	ua = m.NewUploadAdapter("test")
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

func TestAdapterRegButBasicOnly(t *testing.T) {
	cli, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.basictransfersonly": "yes",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")

	assert := assert.New(t)

	m.RegisterNewAdapterFunc("test", Upload, newTestAdapter)
	m.RegisterNewAdapterFunc("test", Download, newTestAdapter)
	// Will still be created if we ask for them
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

	// But list will exclude
	ld := m.GetDownloadAdapterNames()
	assert.Equal([]string{"basic"}, ld)
	lu := m.GetUploadAdapterNames()
	assert.Equal([]string{"basic"}, lu)
}
