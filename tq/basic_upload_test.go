package tq

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicUploadMakeRequestClosesUnauthorizedBody(t *testing.T) {
	srv, open := newUnauthorizedServer(t)

	a := &basicUploadAdapter{newAdapterBase(nil, BasicAdapterName, Upload, nil)}
	a.apiClient = newTestAPIClient(t, srv)

	path := filepath.Join(t.TempDir(), "object")
	require.NoError(t, os.WriteFile(path, []byte("content"), 0644))

	// Each href is distinct and does not contain the OID, so access is
	// cached per object and every request starts with a 401.
	const objects = 5
	for i := range objects {
		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/objects/%d", srv.URL, i), nil)
		require.NoError(t, err)

		res, err := a.makeRequest(&Transfer{Oid: "oid", Path: path, Size: 7}, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		_, err = io.Copy(io.Discard, res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
	}

	requireOnlyKeepAliveConnection(t, open)
}
