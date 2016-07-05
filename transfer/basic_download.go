package transfer

import (
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/localstorage"
	"github.com/github/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

// Adapter for basic HTTP downloads, includes resuming via HTTP Range
type basicDownloadAdapter struct {
	*adapterBase
}

func (a *basicDownloadAdapter) ClearTempStorage() error {
	return os.RemoveAll(a.tempDir())
}

func (a *basicDownloadAdapter) tempDir() string {
	// Must be dedicated to this adapter as deleted by ClearTempStorage
	// Also make local to this repo not global, and separate to localstorage temp,
	// which gets cleared at the end of every invocation
	d := filepath.Join(localstorage.Objects().RootDir, "incomplete")
	if err := os.MkdirAll(d, 0755); err != nil {
		return os.TempDir()
	}
	return d
}

func (a *basicDownloadAdapter) DoTransfer(t *Transfer, cb TransferProgressCallback, authOkFunc func()) error {

	f, fromByte, hashSoFar, err := a.checkResumeDownload(t)
	if err != nil {
		return err
	}
	return a.download(t, cb, authOkFunc, f, fromByte, hashSoFar)
}

// Checks to see if a download can be resumed, and if so returns a non-nil locked file, byte start and hash
func (a *basicDownloadAdapter) checkResumeDownload(t *Transfer) (outFile *os.File, fromByte int64, hashSoFar hash.Hash, e error) {
	// lock the file by opening it for read/write, rather than checking Stat() etc
	// which could be subject to race conditions by other processes
	f, err := os.OpenFile(a.downloadFilename(t), os.O_RDWR, 0644)

	if err != nil {
		// Create a new file instead, must not already exist or error (permissions / race condition)
		newfile, err := os.OpenFile(a.downloadFilename(t), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
		return newfile, 0, nil, err
	}

	// Successfully opened an existing file at this point
	// Read any existing data into hash then return file handle at end
	hash := tools.NewLfsContentHash()
	n, err := io.Copy(hash, f)
	if err != nil {
		f.Close()
		return nil, 0, nil, err
	}
	tracerx.Printf("xfer: Attempting to resume download of %q from byte %d", t.Object.Oid, n)
	return f, n, hash, nil

}

// Create or open a download file for resuming
func (a *basicDownloadAdapter) downloadFilename(t *Transfer) string {
	// Not a temp file since we will be resuming it
	return filepath.Join(a.tempDir(), t.Object.Oid+".tmp")
}

// download starts or resumes and download. Always closes dlFile if non-nil
func (a *basicDownloadAdapter) download(t *Transfer, cb TransferProgressCallback, authOkFunc func(), dlFile *os.File, fromByte int64, hash hash.Hash) error {

	if dlFile != nil {
		// ensure we always close dlFile. Note that this does not conflict with the
		// early close below, as close is idempotent.
		defer dlFile.Close()
	}

	rel, ok := t.Object.Rel("download")
	if !ok {
		return errors.New("Object not found on the server.")
	}

	req, err := httputil.NewHttpRequest("GET", rel.Href, rel.Header)
	if err != nil {
		return err
	}

	if fromByte > 0 {
		if dlFile == nil || hash == nil {
			return fmt.Errorf("Cannot restart %v from %d without a file & hash", t.Object.Oid, fromByte)
		}
		// We could just use a start byte, but since we know the length be specific
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", fromByte, t.Object.Size-1))
	}

	res, err := httputil.DoHttpRequest(req, true)
	if err != nil {
		// Special-case status code 416 () - fall back
		if fromByte > 0 && dlFile != nil && res.StatusCode == 416 {
			tracerx.Printf("xfer: server rejected resume download request for %q from byte %d; re-downloading from start", t.Object.Oid, fromByte)
			dlFile.Close()
			os.Remove(dlFile.Name())
			return a.download(t, cb, authOkFunc, nil, 0, nil)
		}
		return errutil.NewRetriableError(err)
	}
	httputil.LogTransfer("lfs.data.download", res)
	defer res.Body.Close()

	// Range request must return 206 & content range to confirm
	if fromByte > 0 {
		rangeRequestOk := false
		var failReason string
		// check 206 and Content-Range, fall back if either not as expected
		if res.StatusCode == 206 {
			// Probably a successful range request, check Content-Range
			if rangeHdr := res.Header.Get("Content-Range"); rangeHdr != "" {
				regex := regexp.MustCompile(`bytes (\d+)\-.*`)
				match := regex.FindStringSubmatch(rangeHdr)
				if match != nil && len(match) > 1 {
					contentStart, _ := strconv.ParseInt(match[1], 10, 64)
					if contentStart == fromByte {
						rangeRequestOk = true
					} else {
						failReason = fmt.Sprintf("Content-Range start byte incorrect: %s expected %d", match[1], fromByte)
					}
				} else {
					failReason = fmt.Sprintf("badly formatted Content-Range header: %q", rangeHdr)
				}
			} else {
				failReason = "missing Content-Range header in response"
			}
		} else {
			failReason = fmt.Sprintf("expected status code 206, received %d", res.StatusCode)
		}
		if rangeRequestOk {
			tracerx.Printf("xfer: server accepted resume download request: %q from byte %d", t.Object.Oid, fromByte)
			advanceCallbackProgress(cb, t, fromByte)
		} else {
			// Abort resume, perform regular download
			tracerx.Printf("xfer: failed to resume download for %q from byte %d: %s. Re-downloading from start", t.Object.Oid, fromByte, failReason)
			dlFile.Close()
			os.Remove(dlFile.Name())
			if res.StatusCode == 200 {
				// If status code was 200 then server just ignored Range header and
				// sent everything. Don't re-request, use this one from byte 0
				dlFile = nil
				fromByte = 0
				hash = nil
			} else {
				// re-request needed
				return a.download(t, cb, authOkFunc, nil, 0, nil)
			}
		}
	}

	// Signal auth OK on success response, before starting download to free up
	// other workers immediately
	if authOkFunc != nil {
		authOkFunc()
	}

	var hasher *tools.HashingReader
	if fromByte > 0 && hash != nil {
		// pre-load hashing reader with previous content
		hasher = tools.NewHashingReaderPreloadHash(res.Body, hash)
	} else {
		hasher = tools.NewHashingReader(res.Body)
	}

	if dlFile == nil {
		// New file start
		dlFile, err = os.OpenFile(a.downloadFilename(t), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dlFile.Close()
	}
	dlfilename := dlFile.Name()
	// Wrap callback to give name context
	ccb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		if cb != nil {
			return cb(t.Name, totalSize, readSoFar+fromByte, readSinceLast)
		}
		return nil
	}
	written, err := tools.CopyWithCallback(dlFile, hasher, res.ContentLength, ccb)
	if err != nil {
		return fmt.Errorf("cannot write data to tempfile %q: %v", dlfilename, err)
	}
	if err := dlFile.Close(); err != nil {
		return fmt.Errorf("can't close tempfile %q: %v", dlfilename, err)
	}

	if actual := hasher.Hash(); actual != t.Object.Oid {
		return fmt.Errorf("Expected OID %s, got %s after %d bytes written", t.Object.Oid, actual, written)
	}

	return tools.RenameFileCopyPermissions(dlfilename, t.Path)

}

func init() {
	newfunc := func(name string, dir Direction) TransferAdapter {
		switch dir {
		case Download:
			bd := &basicDownloadAdapter{newAdapterBase(name, dir, nil)}
			// self implements impl
			bd.transferImpl = bd
			return bd
		case Upload:
			panic("Should never ask this func to upload")
		}
		return nil
	}
	RegisterNewTransferAdapterFunc(BasicAdapterName, Download, newfunc)
}
