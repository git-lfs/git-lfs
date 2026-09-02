package tq

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicUploadAdapterContentLengthWithChunked(t *testing.T) {
	const content = "hello chunked upload"

	f, err := os.CreateTemp("", "git-lfs-upload-test-*")
	require.Nil(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString(content)
	require.Nil(t, err)
	f.Close()

	tests := []struct {
		name              string
		actionHeaders     map[string]string
		wantContentLength string
	}{
		{
			name:              "chunked: Content-Length header must be absent",
			actionHeaders:     map[string]string{"Transfer-Encoding": "chunked"},
			wantContentLength: "",
		},
		{
			name:              "chunked uppercase: Content-Length header must be absent",
			actionHeaders:     map[string]string{"Transfer-Encoding": "Chunked"},
			wantContentLength: "",
		},
		{
			name:              "chunked in list: Content-Length header must be absent",
			actionHeaders:     map[string]string{"Transfer-Encoding": "gzip, chunked"},
			wantContentLength: "",
		},
		{
			name:              "normal: Content-Length header must equal file size",
			actionHeaders:     map[string]string{},
			wantContentLength: strconv.Itoa(len(content)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedContentLength string
			var requestReceived bool

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestReceived = true
				capturedContentLength = r.Header.Get("Content-Length")
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c := lfsapi.NewClient(lfshttp.NewContext(nil, nil, nil))
			bu := &basicUploadAdapter{newAdapterBase(nil, BasicAdapterName, Upload, nil)}
			bu.transferImpl = bu
			bu.apiClient = c

			transfer := &Transfer{
				Oid:           "abc123def456",
				Size:          int64(len(content)),
				Path:          f.Name(),
				Authenticated: true,
				Actions: map[string]*Action{
					"upload": {
						Href:   srv.URL + "/upload",
						Header: tc.actionHeaders,
					},
				},
			}

			err := bu.DoTransfer(nil, transfer, nil, nil)
			require.Nil(t, err)
			assert.True(t, requestReceived, "expected upload request to reach the server")
			assert.Equal(t, tc.wantContentLength, capturedContentLength)
		})
	}
}
