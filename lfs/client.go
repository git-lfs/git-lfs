package lfs

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

const (
	mediaType = "application/vnd.git-lfs+json; charset=utf-8"
)

// The apiEvent* statuses (and the apiEvent channel) are used by
// UploadQueue to know when it is OK to process uploads concurrently.
const (
	apiEventFail = iota
	apiEventSuccess
)

var (
	lfsMediaTypeRE             = regexp.MustCompile(`\Aapplication/vnd\.git\-lfs\+json(;|\z)`)
	jsonMediaTypeRE            = regexp.MustCompile(`\Aapplication/json(;|\z)`)
	objectRelationDoesNotExist = errors.New("relation does not exist")
	hiddenHeaders              = map[string]bool{
		"Authorization": true,
	}

	defaultErrors = map[int]string{
		400: "Client error: %s",
		401: "Authorization error: %s\nCheck that you have proper access to the repository",
		403: "Authorization error: %s\nCheck that you have proper access to the repository",
		404: "Repository or object not found: %s\nCheck that it exists and that you have proper access to it",
		500: "Server error: %s",
	}

	apiEvent = make(chan int)
)

type ObjectResource struct {
	Oid   string                   `json:"oid,omitempty"`
	Size  int64                    `json:"size,omitempty"`
	Links map[string]*linkRelation `json:"_links,omitempty"`
}

func (o *ObjectResource) CanDownload() bool {
	_, ok := o.Rel("download")
	return ok
}
func (o *ObjectResource) CanUpload() bool {
	_, ok := o.Rel("upload")
	return ok
}

func (o *ObjectResource) NewRequest(ctx *HttpApiContext, relation, method string) (*http.Request, Creds, error) {
	rel, ok := o.Rel(relation)
	if !ok {
		return nil, nil, objectRelationDoesNotExist
	}

	req, creds, err := ctx.newClientRequest(method, rel.Href)
	if err != nil {
		return nil, nil, err
	}

	for h, v := range rel.Header {
		req.Header.Set(h, v)
	}

	return req, creds, nil
}

func (o *ObjectResource) Rel(name string) (*linkRelation, bool) {
	if o.Links == nil {
		return nil, false
	}

	rel, ok := o.Links[name]
	return rel, ok
}

type linkRelation struct {
	Href   string            `json:"href"`
	Header map[string]string `json:"header,omitempty"`
}

type ClientError struct {
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url,omitempty"`
	RequestId        string `json:"request_id,omitempty"`
}

func (e *ClientError) Error() string {
	msg := e.Message
	if len(e.DocumentationUrl) > 0 {
		msg += "\nDocs: " + e.DocumentationUrl
	}
	if len(e.RequestId) > 0 {
		msg += "\nRequest ID: " + e.RequestId
	}
	return msg
}

func Download(oid string) (io.ReadCloser, int64, *WrappedError) {

	ctx := GetApiContext(Config.Endpoint())
	defer ReleaseApiContext(ctx)
	return ctx.Download(oid)
}

type byteCloser struct {
	*bytes.Reader
}

func DownloadCheck(oid string) (*ObjectResource, *WrappedError) {
	ctx := GetApiContext(Config.Endpoint())
	defer ReleaseApiContext(ctx)
	return ctx.DownloadCheck(oid)
}

func DownloadObject(obj *ObjectResource) (io.ReadCloser, int64, *WrappedError) {
	ctx := GetApiContext(Config.Endpoint())
	defer ReleaseApiContext(ctx)
	return ctx.DownloadObject(obj)
}

func (b *byteCloser) Close() error {
	return nil
}

func Batch(objects []*ObjectResource) ([]*ObjectResource, *WrappedError) {
	ctx := GetApiContext(Config.Endpoint())
	defer ReleaseApiContext(ctx)
	return ctx.Batch(objects)
}

func UploadCheck(oidPath string) (*ObjectResource, *WrappedError) {
	oid := filepath.Base(oidPath)

	stat, err := os.Stat(oidPath)
	if err != nil {
		sendApiEvent(apiEventFail)
		return nil, Error(err)
	}

	ctx := GetApiContext(Config.Endpoint())
	defer ReleaseApiContext(ctx)
	return ctx.UploadCheck(oid, stat.Size())
}

func UploadObject(o *ObjectResource, cb CopyCallback) *WrappedError {

	path, err := LocalMediaPath(o.Oid)
	if err != nil {
		return Error(err)
	}

	file, err := os.Open(path)
	if err != nil {
		return Error(err)
	}
	defer file.Close()

	reader := &CallbackReader{
		C:         cb,
		TotalSize: o.Size,
		Reader:    file,
	}

	ctx := GetApiContext(Config.Endpoint())
	defer ReleaseApiContext(ctx)
	return ctx.UploadObject(o, reader)
}

func saveCredentials(creds Creds, res *http.Response) {
	if creds == nil {
		return
	}

	if res.StatusCode < 300 {
		execCreds(creds, "approve")
	} else if res.StatusCode == 401 {
		execCreds(creds, "reject")
	}
}

func getCreds(req *http.Request) (Creds, error) {
	if len(req.Header.Get("Authorization")) > 0 {
		return nil, nil
	}

	apiUrl, err := Config.ObjectUrl("")
	if err != nil {
		return nil, err
	}

	if req.URL.Scheme == apiUrl.Scheme &&
		req.URL.Host == apiUrl.Host {
		creds, err := credentials(req.URL)
		if err != nil {
			return nil, err
		}

		token := fmt.Sprintf("%s:%s", creds["username"], creds["password"])
		auth := "Basic " + base64.URLEncoding.EncodeToString([]byte(token))
		req.Header.Set("Authorization", auth)
		return creds, nil
	}

	return nil, nil
}

func sendApiEvent(event int) {
	select {
	case apiEvent <- event:
	default:
	}
}
