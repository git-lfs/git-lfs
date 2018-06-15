package tq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestAPIBatch(t *testing.T) {
	require.NotNil(t, batchReqSchema, batchReqSchema.Source)
	require.NotNil(t, batchResSchema, batchResSchema.Source)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/objects/batch" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "91", r.Header.Get("Content-Length"))

		bodyLoader, body := gojsonschema.NewReaderLoader(r.Body)
		bReq := &batchRequest{}
		err := json.NewDecoder(body).Decode(bReq)
		r.Body.Close()
		assert.Nil(t, err)
		assertSchema(t, batchReqSchema, bodyLoader)

		assert.EqualValues(t, []string{"basic", "whatev"}, bReq.TransferAdapterNames)
		if assert.Equal(t, 1, len(bReq.Objects)) {
			assert.Equal(t, "a", bReq.Objects[0].Oid)
		}

		w.Header().Set("Content-Type", "application/json")

		writeLoader, resWriter := gojsonschema.NewWriterLoader(w)
		err = json.NewEncoder(resWriter).Encode(&BatchResponse{
			TransferAdapterName: "basic",
			Objects:             bReq.Objects,
		})

		assert.Nil(t, err)
		assertSchema(t, batchResSchema, writeLoader)
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(lfsapi.NewContext(nil, nil, map[string]string{
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
	require.NotNil(t, batchReqSchema, batchReqSchema.Source)
	require.NotNil(t, batchResSchema, batchResSchema.Source)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/objects/batch" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "POST", r.Method)

		bodyLoader, body := gojsonschema.NewReaderLoader(r.Body)
		bReq := &batchRequest{}
		err := json.NewDecoder(body).Decode(bReq)
		r.Body.Close()
		assert.Nil(t, err)
		assertSchema(t, batchReqSchema, bodyLoader)

		assert.Equal(t, 0, len(bReq.TransferAdapterNames))
		if assert.Equal(t, 1, len(bReq.Objects)) {
			assert.Equal(t, "a", bReq.Objects[0].Oid)
		}

		w.Header().Set("Content-Type", "application/json")
		writeLoader, resWriter := gojsonschema.NewWriterLoader(w)
		err = json.NewEncoder(resWriter).Encode(&BatchResponse{
			TransferAdapterName: "basic",
			Objects:             make([]*Transfer, 0),
		})

		assert.Nil(t, err)
		assertSchema(t, batchResSchema, writeLoader)
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(lfsapi.NewContext(nil, nil, map[string]string{
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
	c, err := lfsapi.NewClient(nil)
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

var (
	batchReqSchema *sourcedSchema
	batchResSchema *sourcedSchema
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("getwd error:", err)
		return
	}

	batchReqSchema = getSchema(wd, "schemas/http-batch-request-schema.json")
	batchResSchema = getSchema(wd, "schemas/http-batch-response-schema.json")
}

type sourcedSchema struct {
	Source string
	*gojsonschema.Schema
}

func getSchema(wd, relpath string) *sourcedSchema {
	abspath := filepath.ToSlash(filepath.Join(wd, relpath))
	s, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader(fmt.Sprintf("file:///%s", abspath)))
	if err != nil {
		fmt.Printf("schema load error for %q: %+v\n", relpath, err)
	}
	return &sourcedSchema{Source: relpath, Schema: s}
}

func assertSchema(t *testing.T, schema *sourcedSchema, dataLoader gojsonschema.JSONLoader) {
	res, err := schema.Validate(dataLoader)
	if assert.Nil(t, err) {
		if res.Valid() {
			return
		}

		resErrors := res.Errors()
		valErrors := make([]string, 0, len(resErrors))
		for _, resErr := range resErrors {
			valErrors = append(valErrors, resErr.String())
		}
		t.Errorf("Schema: %s\n%s", schema.Source, strings.Join(valErrors, "\n"))
	}
}
