package lfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tq"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/rubyist/tracerx"
)

func PointerSmudgeToFile(filename string, ptr *Pointer, download bool, manifest *tq.Manifest, cb progress.CopyCallback) error {
	os.MkdirAll(filepath.Dir(filename), 0755)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Could not create working directory file: %v", err)
	}
	defer file.Close()
	if err := PointerSmudge(file, ptr, filename, download, manifest, cb); err != nil {
		if errors.IsDownloadDeclinedError(err) {
			// write placeholder data instead
			file.Seek(0, os.SEEK_SET)
			ptr.Encode(file)
			return err
		} else {
			return fmt.Errorf("Could not write working directory file: %v", err)
		}
	}
	return nil
}

func PointerSmudge(writer io.Writer, ptr *Pointer, workingfile string, download bool, manifest *tq.Manifest, cb progress.CopyCallback) error {
	mediafile, err := LocalMediaPath(ptr.Oid)
	if err != nil {
		return err
	}

	LinkOrCopyFromReference(ptr.Oid, ptr.Size)

	stat, statErr := os.Stat(mediafile)
	if statErr == nil && stat != nil {
		fileSize := stat.Size()
		if fileSize == 0 || fileSize != ptr.Size {
			tracerx.Printf("Removing %s, size %d is invalid", mediafile, fileSize)
			os.RemoveAll(mediafile)
			stat = nil
		}
	}

	if statErr != nil || stat == nil {
		if download {
			err = downloadFile(writer, ptr, workingfile, mediafile, manifest, cb)
		} else {
			return errors.NewDownloadDeclinedError(statErr, "smudge")
		}
	} else {
		err = readLocalFile(writer, ptr, mediafile, workingfile, cb)
	}

	if err != nil {
		return errors.NewSmudgeError(err, ptr.Oid, mediafile)
	}

	return nil
}

type tempAdapterConfig struct {
	concurrentTransfers int
}

func (c tempAdapterConfig) ConcurrentTransfers() int {
	return c.concurrentTransfers
}

func downloadFile(writer io.Writer, ptr *Pointer, workingfile, mediafile string, manifest *tq.Manifest, cb progress.CopyCallback) error {
	fmt.Fprintf(os.Stderr, "Downloading %s (%s)\n", workingfile, pb.FormatBytes(ptr.Size))

	xfers := manifest.GetDownloadAdapterNames()
	obj, adapterName, err := api.BatchSingle(config.Config, &api.ObjectResource{Oid: ptr.Oid, Size: ptr.Size}, "download", xfers)
	if err != nil {
		return errors.Wrapf(err, "Error downloading %s: %s", filepath.Base(mediafile), err)
	}

	if ptr.Size == 0 {
		ptr.Size = obj.Size
	}

	adapter := manifest.NewDownloadAdapter(adapterName)
	var tcb tq.ProgressCallback
	if cb != nil {
		tcb = func(name string, totalSize, readSoFar int64, readSinceLast int) error {
			return cb(totalSize, readSoFar, readSinceLast)
		}
	}
	// Single download
	err = adapter.Begin(tempAdapterConfig{concurrentTransfers: 1}, tcb)
	if err != nil {
		return err
	}
	res := <-adapter.Add(tq.NewTransfer(filepath.Base(workingfile), obj, mediafile))
	adapter.End()

	if res.Error != nil {
		return errors.Wrapf(err, "Error buffering media file: %s", res.Error)
	}

	return readLocalFile(writer, ptr, mediafile, workingfile, nil)
}

func readLocalFile(writer io.Writer, ptr *Pointer, mediafile string, workingfile string, cb progress.CopyCallback) error {
	reader, err := os.Open(mediafile)
	if err != nil {
		return errors.Wrapf(err, "Error opening media file.")
	}
	defer reader.Close()

	if ptr.Size == 0 {
		if stat, _ := os.Stat(mediafile); stat != nil {
			ptr.Size = stat.Size()
		}
	}

	if len(ptr.Extensions) > 0 {
		registeredExts := config.Config.Extensions()
		extensions := make(map[string]config.Extension)
		for _, ptrExt := range ptr.Extensions {
			ext, ok := registeredExts[ptrExt.Name]
			if !ok {
				err := fmt.Errorf("Extension '%s' is not configured.", ptrExt.Name)
				return errors.Wrap(err, "smudge")
			}
			ext.Priority = ptrExt.Priority
			extensions[ext.Name] = ext
		}
		exts, err := config.SortExtensions(extensions)
		if err != nil {
			return errors.Wrap(err, "smudge")
		}

		// pipe extensions in reverse order
		var extsR []config.Extension
		for i := range exts {
			ext := exts[len(exts)-1-i]
			extsR = append(extsR, ext)
		}

		request := &pipeRequest{"smudge", reader, workingfile, extsR}

		response, err := pipeExtensions(request)
		if err != nil {
			return errors.Wrap(err, "smudge")
		}

		actualExts := make(map[string]*pipeExtResult)
		for _, result := range response.results {
			actualExts[result.name] = result
		}

		// verify name, order, and oids
		oid := response.results[0].oidIn
		if ptr.Oid != oid {
			err = fmt.Errorf("Actual oid %s during smudge does not match expected %s", oid, ptr.Oid)
			return errors.Wrap(err, "smudge")
		}

		for _, expected := range ptr.Extensions {
			actual := actualExts[expected.Name]
			if actual.name != expected.Name {
				err = fmt.Errorf("Actual extension name '%s' does not match expected '%s'", actual.name, expected.Name)
				return errors.Wrap(err, "smudge")
			}
			if actual.oidOut != expected.Oid {
				err = fmt.Errorf("Actual oid %s for extension '%s' does not match expected %s", actual.oidOut, expected.Name, expected.Oid)
				return errors.Wrap(err, "smudge")
			}
		}

		// setup reader
		reader, err = os.Open(response.file.Name())
		if err != nil {
			return errors.Wrapf(err, "Error opening smudged file: %s", err)
		}
		defer reader.Close()
	}

	_, err = tools.CopyWithCallback(writer, reader, ptr.Size, cb)
	if err != nil {
		return errors.Wrapf(err, "Error reading from media file: %s", err)
	}

	return nil
}
