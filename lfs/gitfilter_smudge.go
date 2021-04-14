package lfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tools/humanize"
	"github.com/git-lfs/git-lfs/tq"
	"github.com/rubyist/tracerx"
)

func (f *GitFilter) SmudgeToFile(filename string, ptr *Pointer, download bool, manifest *tq.Manifest, cb tools.CopyCallback) error {
	tools.MkdirAll(filepath.Dir(filename), f.cfg)

	if stat, _ := os.Stat(filename); stat != nil && stat.Mode()&0200 == 0 {
		if err := os.Chmod(filename, stat.Mode()|0200); err != nil {
			return errors.Wrap(err,
				"Could not restore write permission")
		}

		// When we're done, return the file back to its normal
		// permission bits.
		defer os.Chmod(filename, stat.Mode())
	}

	abs, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("could not produce absolute path for %q", filename)
	}

	file, err := os.Create(abs)
	if err != nil {
		return fmt.Errorf("could not create working directory file: %v", err)
	}
	defer file.Close()
	if _, err := f.Smudge(file, ptr, filename, download, manifest, cb); err != nil {
		if errors.IsDownloadDeclinedError(err) {
			// write placeholder data instead
			file.Seek(0, io.SeekStart)
			ptr.Encode(file)
			return err
		} else {
			return fmt.Errorf("could not write working directory file: %v", err)
		}
	}
	return nil
}

func (f *GitFilter) Smudge(writer io.Writer, ptr *Pointer, workingfile string, download bool, manifest *tq.Manifest, cb tools.CopyCallback) (int64, error) {
	mediafile, err := f.ObjectPath(ptr.Oid)
	if err != nil {
		return 0, err
	}

	LinkOrCopyFromReference(f.cfg, ptr.Oid, ptr.Size)

	stat, statErr := os.Stat(mediafile)
	if statErr == nil && stat != nil {
		fileSize := stat.Size()
		if fileSize != ptr.Size {
			tracerx.Printf("Removing %s, size %d is invalid", mediafile, fileSize)
			os.RemoveAll(mediafile)
			stat = nil
		}
	}

	var n int64

	if statErr != nil || stat == nil {
		if download {
			n, err = f.downloadFile(writer, ptr, workingfile, mediafile, manifest, cb)
		} else {
			return 0, errors.NewDownloadDeclinedError(statErr, "smudge")
		}
	} else {
		n, err = f.readLocalFile(writer, ptr, mediafile, workingfile, cb)
	}

	if err != nil {
		return 0, errors.NewSmudgeError(err, ptr.Oid, mediafile)
	}

	return n, nil
}

func (f *GitFilter) downloadFile(writer io.Writer, ptr *Pointer, workingfile, mediafile string, manifest *tq.Manifest, cb tools.CopyCallback) (int64, error) {
	fmt.Fprintf(os.Stderr, "Downloading %s (%s)\n", workingfile, humanize.FormatBytes(uint64(ptr.Size)))

	// NOTE: if given, "cb" is a tools.CopyCallback which writes updates
	// to the logpath specified by GIT_LFS_PROGRESS.
	//
	// Either way, forward it into the *tq.TransferQueue so that updates are
	// sent over correctly.

	q := tq.NewTransferQueue(tq.Download, manifest, f.cfg.Remote(),
		tq.WithProgressCallback(cb),
		tq.RemoteRef(f.RemoteRef()),
	)
	q.Add(filepath.Base(workingfile), mediafile, ptr.Oid, ptr.Size, false, nil)
	q.Wait()

	if errs := q.Errors(); len(errs) > 0 {
		var multiErr error
		for _, e := range errs {
			if multiErr != nil {
				multiErr = fmt.Errorf("%v\n%v", multiErr, e)
			} else {
				multiErr = e
			}
		}

		return 0, errors.Wrapf(multiErr, "Error downloading %s (%s)", workingfile, ptr.Oid)
	}

	return f.readLocalFile(writer, ptr, mediafile, workingfile, nil)
}

func (f *GitFilter) readLocalFile(writer io.Writer, ptr *Pointer, mediafile string, workingfile string, cb tools.CopyCallback) (int64, error) {
	reader, err := tools.RobustOpen(mediafile)
	if err != nil {
		return 0, errors.Wrapf(err, "error opening media file")
	}
	defer reader.Close()

	if ptr.Size == 0 {
		if stat, _ := os.Stat(mediafile); stat != nil {
			ptr.Size = stat.Size()
		}
	}

	if len(ptr.Extensions) > 0 {
		registeredExts := f.cfg.Extensions()
		extensions := make(map[string]config.Extension)
		for _, ptrExt := range ptr.Extensions {
			ext, ok := registeredExts[ptrExt.Name]
			if !ok {
				err := fmt.Errorf("extension '%s' is not configured", ptrExt.Name)
				return 0, errors.Wrap(err, "smudge")
			}
			ext.Priority = ptrExt.Priority
			extensions[ext.Name] = ext
		}
		exts, err := config.SortExtensions(extensions)
		if err != nil {
			return 0, errors.Wrap(err, "smudge")
		}

		// pipe extensions in reverse order
		var extsR []config.Extension
		for i := range exts {
			ext := exts[len(exts)-1-i]
			extsR = append(extsR, ext)
		}

		request := &pipeRequest{"smudge", reader, workingfile, extsR}

		response, err := pipeExtensions(f.cfg, request)
		if err != nil {
			return 0, errors.Wrap(err, "smudge")
		}

		actualExts := make(map[string]*pipeExtResult)
		for _, result := range response.results {
			actualExts[result.name] = result
		}

		// verify name, order, and oids
		oid := response.results[0].oidIn
		if ptr.Oid != oid {
			err = fmt.Errorf("actual oid %s during smudge does not match expected %s", oid, ptr.Oid)
			return 0, errors.Wrap(err, "smudge")
		}

		for _, expected := range ptr.Extensions {
			actual := actualExts[expected.Name]
			if actual.name != expected.Name {
				err = fmt.Errorf("actual extension name '%s' does not match expected '%s'", actual.name, expected.Name)
				return 0, errors.Wrap(err, "smudge")
			}
			if actual.oidOut != expected.Oid {
				err = fmt.Errorf("actual oid %s for extension '%s' does not match expected %s", actual.oidOut, expected.Name, expected.Oid)
				return 0, errors.Wrap(err, "smudge")
			}
		}

		// setup reader
		reader, err = os.Open(response.file.Name())
		if err != nil {
			return 0, errors.Wrapf(err, "Error opening smudged file: %s", err)
		}
		defer reader.Close()
	}

	n, err := tools.CopyWithCallback(writer, reader, ptr.Size, cb)
	if err != nil {
		return n, errors.Wrapf(err, "Error reading from media file: %s", err)
	}

	return n, nil
}
