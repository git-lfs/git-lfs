package transfer

import (
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/tools"
)

const (
	HttpRangeAdapterName = "http-range"
)

// Download adapter that can resume downloads using HTTP Range headers
type httpRangeAdapter struct {
	*adapterBase
}

func (a *httpRangeAdapter) ClearTempStorage() error {
	return os.RemoveAll(a.tempDir())
}

func (a *httpRangeAdapter) tempDir() string {
	// Must be dedicated to this adapter as deleted by ClearTempStorage
	d := filepath.Join(os.TempDir(), "git-lfs-http-range-temp")
	if err := os.MkdirAll(d, 0755); err != nil {
		return os.TempDir()
	}
	return d
}

func (a *httpRangeAdapter) DoTransfer(t *Transfer, cb TransferProgressCallback, authOkFunc func()) error {

	f, fromByte, hashSoFar, err := a.checkResumeDownload(t)
	if err != nil {
		return err
	}
	return a.download(t, cb, authOkFunc, f, fromByte, hashSoFar)
}

// Checks to see if a download can be resumed, and if so returns a non-nil locked file, byte start and hash
func (a *httpRangeAdapter) checkResumeDownload(t *Transfer) (outFile *os.File, fromByte int64, hashSoFar hash.Hash, e error) {
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
	return f, n, hash, nil

}

// Create or open a download file for resuming
func (a *httpRangeAdapter) downloadFilename(t *Transfer) string {
	// Not a temp file since we will be resuming it
	return filepath.Join(a.tempDir(), t.Object.Oid+".tmp")
}

// download starts or resumes and download. Always closes dlFile if non-nil
func (a *httpRangeAdapter) download(t *Transfer, cb TransferProgressCallback, authOkFunc func(), dlFile *os.File, fromByte int64, hash hash.Hash) error {

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
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", fromByte, t.Object.Size))
	}

	res, err := httputil.DoHttpRequest(req, true)
	if err != nil {
		return errutil.NewRetriableError(err)
	}
	httputil.LogTransfer("lfs.data.download", res)
	defer res.Body.Close()

	// Range request must return 206 to confirm
	if fromByte > 0 {
		if res.StatusCode == 206 {
			// Successful range request
			// Advance progress callback; must split into max int sizes though
			const maxInt = int(^uint(0) >> 1)
			for read := int64(0); read < fromByte; {
				remainder := fromByte - read
				if remainder > int64(maxInt) {
					read += int64(maxInt)
					cb(t.Name, t.Object.Size, read, maxInt)
				} else {
					read += remainder
					cb(t.Name, t.Object.Size, read, int(remainder))
				}

			}
		} else {
			// Abort resume, perform regular download
			dlFile.Close()
			os.Remove(dlFile.Name())
			return a.download(t, cb, authOkFunc, nil, 0, nil)
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
		dlFile, err := os.OpenFile(a.downloadFilename(t), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

	// Notice that on failure we do not delete the partially downloaded file.
	// Instead we will resume next time

	return tools.RenameFileCopyPermissions(dlfilename, t.Path)

}

func init() {
	newfunc := func(name string, dir Direction) TransferAdapter {
		switch dir {
		case Download:
			hd := &httpRangeAdapter{newAdapterBase(name, dir, nil)}
			// self implements impl
			hd.transferImpl = hd
			return hd
		case Upload:
			panic("Should never ask a HTTP Range adapter to upload")
		}
		return nil
	}
	RegisterNewTransferAdapterFunc(HttpRangeAdapterName, Download, newfunc)
}
