package main

import (
	"fmt"
	"github.com/github/git-lfs/lfs"
	"github.com/technoweenie/go-contentaddressable"
	"io"
	"os"
	"path/filepath"
)

func upload(req *lfs.JsonRequest, in io.Reader, out io.Writer, config *Config, path string) *lfs.JsonResponse {
	upreq := lfs.UploadRequest{}
	err := lfs.ExtractStructFromJsonRawMessage(req.Params, &upreq)
	if err != nil {
		return lfs.NewJsonErrorResponse(req.Id, err.Error())
	}
	startresult := lfs.UploadResponse{}
	startresult.OkToSend = true
	// Send start response immediately
	resp, err := lfs.NewJsonResponse(req.Id, startresult)
	if err != nil {
		return lfs.NewJsonErrorResponse(req.Id, err.Error())
	}
	err = sendResponse(resp, out)
	if err != nil {
		return lfs.NewJsonErrorResponse(req.Id, err.Error())
	}
	// Next from client should be byte stream of exactly the stated number of bytes
	// Write to temporary file then move to final on success
	filename, err := mediaPath(upreq.Oid, config)
	if err != nil {
		return lfs.NewJsonErrorResponse(req.Id, fmt.Sprintf("Error determining media path. %v", err))
	}

	// Now open temp file to write to
	mediaFile, err := contentaddressable.NewFile(filename)
	if err != nil {
		return lfs.NewJsonErrorResponse(req.Id, fmt.Sprintf("Error opening media file buffer. %v", err))
	}

	n, err := io.CopyN(mediaFile, in, upreq.Size)
	if err == nil {
		err = mediaFile.Accept()
	}
	mediaFile.Close()
	if err != nil {
		return lfs.NewJsonErrorResponse(req.Id, fmt.Sprintf("Problem uploading data: %v", err.Error()))
	} else if n != upreq.Size {
		return lfs.NewJsonErrorResponse(req.Id, fmt.Sprintf("Received wrong number of bytes %d (expected %d)", n, upreq.Size))
	}

	receivedresult := lfs.UploadCompleteResponse{}
	receivedresult.ReceivedOk = true
	resp, _ = lfs.NewJsonResponse(req.Id, receivedresult)

	return resp

}

func downloadInfo(req *lfs.JsonRequest, in io.Reader, out io.Writer, config *Config, path string) *lfs.JsonResponse {
	// TODO
	return nil
}
func download(req *lfs.JsonRequest, in io.Reader, out io.Writer, config *Config, path string) *lfs.JsonResponse {
	// TODO
	return nil
}

// Store in the same structure as client, just under BasePath
func mediaPath(sha string, config *Config) (string, error) {
	path := filepath.Join(config.BasePath, sha[0:2], sha[2:4])
	if err := os.MkdirAll(path, 0744); err != nil {
		return "", fmt.Errorf("Error trying to create local media directory in '%s': %s", path, err)
	}
	return filepath.Join(path, sha), nil
}
