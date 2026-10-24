package lfs

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/stretchr/testify/require"
)

type interruptedCleanReader struct {
	err error
}

func (r interruptedCleanReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestGitFilterCleanRemovesPartialFileOnReadError(t *testing.T) {
	dir := t.TempDir()
	cfg := config.NewFrom(config.Values{Git: map[string][]string{
		"lfs.storage": {dir},
	}})
	filter := NewGitFilter(cfg)
	readErr := errors.New("interrupted input")
	reader := io.MultiReader(
		bytes.NewReader(bytes.Repeat([]byte("x"), 2*BlobSizeCutoff)),
		interruptedCleanReader{readErr},
	)

	asset, err := filter.Clean(reader, "interrupted.bin", -1, nil)
	require.ErrorIs(t, err, readErr)
	require.Nil(t, asset)
	files, err := os.ReadDir(cfg.TempDir())
	require.NoError(t, err)
	require.Empty(t, files)
}

func TestGitFilterCleanKeepsSuccessfulFile(t *testing.T) {
	dir := t.TempDir()
	cfg := config.NewFrom(config.Values{Git: map[string][]string{
		"lfs.storage": {dir},
	}})
	filter := NewGitFilter(cfg)
	data := bytes.Repeat([]byte("x"), 2*BlobSizeCutoff)

	asset, err := filter.Clean(bytes.NewReader(data), "success.bin", int64(len(data)), nil)
	require.NoError(t, err)
	require.EqualValues(t, len(data), asset.Size)
	t.Cleanup(func() { asset.Teardown() })
	_, err = os.Stat(asset.Filename)
	require.NoError(t, err)
}
