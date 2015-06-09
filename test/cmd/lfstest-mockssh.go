package main

import (
	"../../lfs"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// This is a fake SSH client which lets us test SSH behaviour in integration tests
// without implementing a whole sshd server locally

var (
	url             string
	storageBasePath string
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Expected 3 arguments, got %v\n", os.Args)
		os.Exit(1)
	}

	// We won't really this
	url = os.Args[1]
	repoPath := os.Args[2]

	// We need to store in the filesystem because multiple calls will be made to fake ssh
	storageBasePath = filepath.Join(os.Getenv("LFSTEST_DIR"), "fakessh", repoPath)

	Serve()
}

// Pretend to be talking to a server
func Serve() {

	// Even though we're not talking to a server and are just a fake ssh client,
	// we still have a loop to process since a single SSH command can be used
	// for multiple operations.

	// Read a request
	rdr := bufio.NewReader(os.Stdin)
	for {
		jsonbytes, err := rdr.ReadBytes(byte(0))
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(fmt.Sprintf("lfstest-fakessh: unable to read from client: %v", err.Error()))
		}
		// slice off the terminator
		jsonbytes = jsonbytes[:len(jsonbytes)-1]
		var req lfs.JsonRequest
		err = json.Unmarshal(jsonbytes, &req)
		if err != nil {
			panic(fmt.Sprintf("lfstest-fakessh: unable to unmarshal json request from client:%v", string(jsonbytes)))
		}

		var resp *lfs.JsonResponse
		switch req.Method {
		case "UploadCheck":
			upreq := lfs.UploadRequest{}
			lfs.ExtractStructFromJsonRawMessage(req.Params, &upreq)
			startresult := lfs.UploadResponse{}
			startresult.OkToSend = fileExists(mediaPath(upreq.Oid))
			resp, err = lfs.NewJsonResponse(req.Id, startresult)
			if err != nil {
				panic("lfstest-fakessh: unable to create response")
			}

		case "Upload":
			upreq := lfs.UploadRequest{}
			lfs.ExtractStructFromJsonRawMessage(req.Params, &upreq)
			startresult := lfs.UploadResponse{}
			startresult.OkToSend = true
			// Send start response immediately
			resp, err = lfs.NewJsonResponse(req.Id, startresult)
			if err != nil {
				panic("lfstest-fakessh: unable to create response")
			}
			responseBytes, err := json.Marshal(resp)
			if err != nil {
				panic("lfstest-fakessh: unable to marshal response")
			}
			// null terminate response
			responseBytes = append(responseBytes, byte(0))
			os.Stdout.Write(responseBytes)
			// Next should by byte stream
			// Must read from buffered reader since bytes may have been read already
			receivedresult := lfs.UploadCompleteResponse{}
			receivedresult.ReceivedOk = true
			var receiveerr error
			file := mediaPath(upreq.Oid)
			f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
			if err != nil {
				panic(fmt.Sprintf("lfstest-fakessh: error opening file %v to write: %v", file, err))
			}
			c, err := io.CopyN(f, rdr, upreq.Size)
			f.Close() // close manually rather than on defer since one method
			if err != nil {
				receivedresult.ReceivedOk = false
				receiveerr = fmt.Errorf("lfstest-fakessh: unable to read data: %v", err.Error())
				break
			} else if c != upreq.Size {
				receivedresult.ReceivedOk = false
				receiveerr = fmt.Errorf("lfstest-fakessh: read wrong number of bytes for %v, expected %d, got %d", file, upreq.Size, c)
			}

			// After we've read all the content bytes, send received response
			if receiveerr != nil {
				resp = lfs.NewJsonErrorResponse(req.Id, receiveerr.Error())
			} else {
				resp, _ = lfs.NewJsonResponse(req.Id, receivedresult)
			}
		case "DownloadCheck":
			downreq := lfs.DownloadCheckRequest{}
			lfs.ExtractStructFromJsonRawMessage(req.Params, &downreq)
			result := lfs.DownloadCheckResponse{}
			file := mediaPath(downreq.Oid)
			s, err := os.Stat(file)
			if err == nil {
				result.Size = s.Size()
			} else {
				result.Size = -1
			}
			resp, err = lfs.NewJsonResponse(req.Id, result)
			if err != nil {
				panic("lfstest-fakessh: unable to create response")
			}
		case "Download":
			// Can't return any error responses here (byte stream response only), have to just fail
			downreq := lfs.DownloadRequest{}
			lfs.ExtractStructFromJsonRawMessage(req.Params, &downreq)
			// there is no response to this
			file := mediaPath(downreq.Oid)
			f, err := os.OpenFile(file, os.O_RDONLY, 0644)
			if err != nil {
				panic(fmt.Sprintf("lfstest-fakessh: error opening file %v to read: %v", file, err))
			}
			c, err := io.Copy(os.Stdout, f)
			f.Close() // close manually rather than on defer since one method
			if err != nil {
				panic(fmt.Sprintf("lfstest-fakessh: error reading from %v: %v", file, err))
			} else if c != downreq.Size {
				panic(fmt.Sprintf("lfstest-fakessh: read wrong number of bytes for %v, expected %d, got %d", file, downreq.Size, c))
			}

		case "Batch":
			batchreq := lfs.BatchRequest{}
			lfs.ExtractStructFromJsonRawMessage(req.Params, &batchreq)
			result := lfs.BatchResponse{}
			for _, o := range batchreq.Objects {
				file := mediaPath(o.Oid)
				s, err := os.Stat(file)
				if err == nil {
					result.Results = append(result.Results, lfs.BatchResponseObject{o.Oid, "download", s.Size()})
				} else {
					result.Results = append(result.Results, lfs.BatchResponseObject{o.Oid, "upload", o.Size})
				}
			}
			resp, err = lfs.NewJsonResponse(req.Id, result)
			if err != nil {
				panic("lfstest-fakessh: unable to create response")
			}
		case "Exit":
			break
		default:
			resp = lfs.NewJsonErrorResponse(req.Id, fmt.Sprintf("Unknown method %v", req.Method))
		}
		if resp != nil {
			responseBytes, err := json.Marshal(resp)
			if err != nil {
				panic("lfstest-fakessh: unable to marshal response")
			}

			// null terminate response
			responseBytes = append(responseBytes, byte(0))
			os.Stdout.Write(responseBytes)
		}
	}
}

func mediaPath(sha string) string {
	abspath := filepath.Join(storageBasePath, sha[0:2], sha[2:4])
	os.MkdirAll(abspath, 0755)
	return filepath.Join(abspath, sha)
}

func fileExists(path string) bool {
	s, err := os.Stat(path)
	if err == nil {
		return !s.IsDir()
	}
	return false
}
