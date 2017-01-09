package tq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIBatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/objects/batch" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "POST", r.Method)

		bReq := &batchRequest{}
		err := json.NewDecoder(r.Body).Decode(bReq)
		r.Body.Close()
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"basic", "whatev"}, bReq.TransferAdapterNames)
		if assert.Equal(t, 1, len(bReq.Objects)) {
			assert.Equal(t, "a", bReq.Objects[0].Oid)
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(&BatchResponse{
			TransferAdapterName: "basic",
			Objects:             bReq.Objects,
		})
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.url": srv.URL + "/api",
	}))
	require.Nil(t, err)

	tqc := &tqClient{Client: c}
	bReq := &batchRequest{
		TransferAdapterNames: []string{"basic", "whatev"},
		Objects: []*Transfer{
			&Transfer{Oid: "a", Size: 1},
		},
	}
	bRes, err := tqc.Batch("remote", bReq)
	require.Nil(t, err)
	assert.Equal(t, "basic", bRes.TransferAdapterName)
	if assert.Equal(t, 1, len(bRes.Objects)) {
		assert.Equal(t, "a", bRes.Objects[0].Oid)
	}
}

func TestAPIBatchOnlyBasic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/objects/batch" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "POST", r.Method)

		bReq := &batchRequest{}
		err := json.NewDecoder(r.Body).Decode(bReq)
		r.Body.Close()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(bReq.TransferAdapterNames))
		if assert.Equal(t, 1, len(bReq.Objects)) {
			assert.Equal(t, "a", bReq.Objects[0].Oid)
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(&BatchResponse{
			TransferAdapterName: "basic",
		})
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.url": srv.URL + "/api",
	}))
	require.Nil(t, err)

	tqc := &tqClient{Client: c}
	bReq := &batchRequest{
		TransferAdapterNames: []string{"basic"},
		Objects: []*Transfer{
			&Transfer{Oid: "a", Size: 1},
		},
	}
	bRes, err := tqc.Batch("remote", bReq)
	require.Nil(t, err)
	assert.Equal(t, "basic", bRes.TransferAdapterName)
}

func TestAPIBatchEmptyObjects(t *testing.T) {
	c, err := lfsapi.NewClient(nil, nil)
	require.Nil(t, err)

	tqc := &tqClient{Client: c}
	bReq := &batchRequest{
		TransferAdapterNames: []string{"basic", "whatev"},
	}
	bRes, err := tqc.Batch("remote", bReq)
	require.Nil(t, err)
	assert.Equal(t, "", bRes.TransferAdapterName)
	assert.Equal(t, 0, len(bRes.Objects))
}
