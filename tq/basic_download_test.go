package tq

import (
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/v3/creds"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type staticCredentialHelper struct{}

func (staticCredentialHelper) Fill(input creds.Creds) (creds.Creds, error) {
	output := make(creds.Creds)
	maps.Copy(output, input)
	output["username"] = []string{"user"}
	output["password"] = []string{"pass"}
	return output, nil
}

func (staticCredentialHelper) Approve(creds.Creds) error { return nil }
func (staticCredentialHelper) Reject(creds.Creds) error  { return nil }

// newUnauthorizedServer starts a server that answers every request without an
// Authorization header with a non-JSON 401, and counts its open connections.
// An unclosed 401 body keeps its connection open, which leaks one file
// descriptor on the client per request.
func newUnauthorizedServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	var open atomic.Int32

	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		io.Copy(io.Discard, req.Body)
		if len(req.Header.Get("Authorization")) == 0 {
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("Lfs-Authenticate", "Basic")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "<html>unauthorized</html>")
			return
		}
		fmt.Fprint(w, "content")
	}))
	srv.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			open.Add(1)
		case http.StateClosed:
			open.Add(-1)
		}
	}
	srv.Start()
	t.Cleanup(srv.Close)

	return srv, &open
}

func newTestAPIClient(t *testing.T, srv *httptest.Server) *lfsapi.Client {
	c := lfsapi.NewClient(lfshttp.NewContext(git.NewReadOnlyConfig("", ""),
		nil, map[string]string{
			"lfs.url": srv.URL + "/repo/lfs",
		},
	))
	c.Credentials = staticCredentialHelper{}
	t.Cleanup(func() { c.Close() })

	return c
}

// requireOnlyKeepAliveConnection checks that every connection except the
// keep-alive one of the last response has been closed.
func requireOnlyKeepAliveConnection(t *testing.T, open *atomic.Int32) {
	require.Eventually(t, func() bool { return open.Load() <= 1 },
		5*time.Second, 10*time.Millisecond)
}

func TestBasicDownloadMakeRequestClosesUnauthorizedBody(t *testing.T) {
	srv, open := newUnauthorizedServer(t)

	a := &basicDownloadAdapter{newAdapterBase(nil, BasicAdapterName, Download, nil)}
	a.apiClient = newTestAPIClient(t, srv)

	// Each href is distinct and does not contain the OID, so access is
	// cached per object and every request starts with a 401.
	const objects = 5
	for i := range objects {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/objects/%d", srv.URL, i), nil)
		require.NoError(t, err)

		res, err := a.makeRequest(&Transfer{Oid: "oid"}, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		_, err = io.Copy(io.Discard, res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
	}

	requireOnlyKeepAliveConnection(t, open)
}
