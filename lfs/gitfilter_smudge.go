package lfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tools/humanize"
	"github.com/git-lfs/git-lfs/v3/tq"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

func (f *GitFilter) SmudgeToFile(path string, ptr *WrappedPointer, download bool, manifest tq.Manifest, cb tools.CopyCallback) error {
	// When no pointer file exists on disk, we should use the permissions
	// defined for the file in Git, since the executable mode may be set.
	// However, to conform with our legacy behaviour, we do not do this
	// at present.
	var mode os.FileMode = 0666
	if stat, _ := os.Lstat(path); stat != nil && stat.Mode().IsRegular() {
		if ptr.Size == 0 && stat.Size() == 0 {
			return nil
		}

		mode = stat.Mode().Perm()
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return errors.New(tr.Tr.Get("could not produce absolute path for %q", path))
	}

	// Write to a temp file in the same directory, then rename atomically.
	// This avoids the Windows Server 2019 race where os.Remove() leaves
	// a pending-deletion entry that blocks os.OpenFile(O_EXCL).
	tmp, err := os.CreateTemp(filepath.Dir(abs), ".lfs-*")
	if err != nil {
		return errors.Wrap(err, tr.Tr.Get("could not create working directory file %q", path))
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName) // no-op if rename succeeded
	}()

	if err := tmp.Chmod(mode); err != nil {
		return errors.Wrap(err, tr.Tr.Get("could not set permissions on temporary file"))
	}

	if _, err := f.Smudge(tmp, ptr.Pointer, ptr.Name, download, manifest, cb); err != nil {
		if errors.IsDownloadDeclinedError(err) {
			// write placeholder data instead
			tmp.Seek(0, io.SeekStart)
			ptr.Encode(tmp)
			if closeErr := tmp.Close(); closeErr == nil {
				os.Rename(tmpName, abs)
			}
			return err
		}
		return errors.New(tr.Tr.Get("could not write working directory file: %v", err))
	}

	if err := tmp.Close(); err != nil {
		return errors.Wrap(err, tr.Tr.Get("could not close temporary file"))
	}

	// Remove symlinks/irregular entries before rename; regular files are
	// replaced atomically by os.Rename without needing a prior Remove.
	if lstat, _ := os.Lstat(abs); lstat != nil {
		if lstat.Mode().IsRegular() {
			// MoveFileExW(MOVEFILE_REPLACE_EXISTING) fails on Windows if the
			// destination is read-only, as may be the case for locked LFS files.
			if lstat.Mode().Perm()&0200 == 0 {
				if err := os.Chmod(abs, lstat.Mode().Perm()|0200); err != nil {
					return errors.Wrap(err, tr.Tr.Get("could not make working directory file writable %q", path))
				}
			}
		} else {
			if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
				return errors.Wrap(err, tr.Tr.Get("could not remove working directory file %q", path))
			}
		}
	}

	if err := os.Rename(tmpName, abs); err != nil {
		return errors.Wrap(err, tr.Tr.Get("could not replace working directory file %q", path))
	}
	return nil
}

func (f *GitFilter) Smudge(writer io.Writer, ptr *Pointer, workingfile string, download bool, manifest tq.Manifest, cb tools.CopyCallback) (int64, error) {
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

	if ptr.Size == 0 {
		return 0, nil
	} else if statErr != nil || stat == nil {
		if download {
			n, err = f.downloadFile(writer, ptr, workingfile, mediafile, manifest, cb)

			// In case of a cherry-pick the newly created commit is likely not yet
			// be found in the history of a remote branch. Thus, the first attempt might fail.
			if err != nil && f.cfg.SearchAllRemotesEnabled() {
				tracerx.Printf("git: smudge: default remote failed. searching alternate remotes")
				n, err = f.downloadFileFallBack(writer, ptr, workingfile, mediafile, manifest, cb)
			}

		} else {
			return 0, errors.NewDownloadDeclinedError(statErr, tr.Tr.Get("smudge filter"))
		}
	} else {
		n, err = f.readLocalFile(writer, ptr, mediafile, workingfile, cb)
	}

	if err != nil {
		return 0, errors.NewSmudgeError(err, ptr.Oid, mediafile)
	}

	return n, nil
}

func (f *GitFilter) downloadFile(writer io.Writer, ptr *Pointer, workingfile, mediafile string, manifest tq.Manifest, cb tools.CopyCallback) (int64, error) {
	fmt.Fprintln(os.Stderr, tr.Tr.Get("Downloading %s (%s)", workingfile, humanize.FormatBytes(uint64(ptr.Size))))

	// NOTE: if given, "cb" is a tools.CopyCallback which writes updates
	// to the logpath specified by GIT_LFS_PROGRESS.
	//
	// Either way, forward it into the *tq.TransferQueue so that updates are
	// sent over correctly.

	q := tq.NewTransferQueue(tq.Download, manifest, f.cfg.Remote(),
		tq.WithProgressCallback(cb),
		tq.RemoteRef(f.RemoteRef()),
		tq.WithBatchSize(f.cfg.TransferBatchSize()),
	)
	q.Add(filepath.Base(workingfile), mediafile, ptr.Oid, ptr.Size, false, nil)
	q.Wait()

	if errs := q.Errors(); len(errs) > 0 {
		return 0, errors.Wrap(errors.Join(errs...), tr.Tr.Get("Error downloading %s (%s)", workingfile, ptr.Oid))
	}

	return f.readLocalFile(writer, ptr, mediafile, workingfile, nil)
}

func (f *GitFilter) downloadFileFallBack(writer io.Writer, ptr *Pointer, workingfile, mediafile string, manifest tq.Manifest, cb tools.CopyCallback) (int64, error) {
	// Attempt to find the LFS objects in all currently registered remotes.
	// When a valid remote is found, this remote is taken persistent for
	// future attempts within downloadFile(). In best case, the ordinary
	// call to downloadFile will then succeed for the rest of files,
	// otherwise this function will again search for a valid remote as fallback.
	remotes := f.cfg.Remotes()
	for index, remote := range remotes {
		q := tq.NewTransferQueue(tq.Download, manifest, remote,
			tq.WithProgressCallback(cb),
			tq.RemoteRef(f.RemoteRef()),
			tq.WithBatchSize(f.cfg.TransferBatchSize()),
		)
		q.Add(filepath.Base(workingfile), mediafile, ptr.Oid, ptr.Size, false, nil)
		q.Wait()

		if errs := q.Errors(); len(errs) > 0 {
			wrappedError := errors.Wrap(errors.Join(errs...), tr.Tr.Get("Error downloading %s (%s)", workingfile, ptr.Oid))
			if index >= len(remotes)-1 {
				return 0, wrappedError
			} else {
				tracerx.Printf("git: download: remote failed %s %s", remote, wrappedError)
			}
		} else {
			// Set the remote persistent through all the operation as we found a valid one.
			// This prevents multiple trial and error searches.
			f.cfg.SetRemote(remote)
			return f.readLocalFile(writer, ptr, mediafile, workingfile, nil)
		}
	}
	return 0, errors.Wrap(errors.New(tr.Tr.Get("No known remotes")), tr.Tr.Get("Error downloading %s (%s)", workingfile, ptr.Oid))
}

func (f *GitFilter) readLocalFile(writer io.Writer, ptr *Pointer, mediafile string, workingfile string, cb tools.CopyCallback) (int64, error) {
	reader, err := tools.RobustOpen(mediafile)
	if err != nil {
		return 0, errors.Wrap(err, tr.Tr.Get("error opening media file"))
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
				err := errors.New(tr.Tr.Get("extension '%s' is not configured", ptrExt.Name))
				return 0, errors.Wrap(err, tr.Tr.Get("smudge filter"))
			}
			ext.Priority = ptrExt.Priority
			extensions[ext.Name] = ext
		}
		exts, err := config.SortExtensions(extensions)
		if err != nil {
			return 0, errors.Wrap(err, tr.Tr.Get("smudge filter"))
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
			return 0, errors.Wrap(err, tr.Tr.Get("smudge filter"))
		}

		actualExts := make(map[string]*pipeExtResult)
		for _, result := range response.results {
			actualExts[result.name] = result
		}

		// verify name, order, and oids
		oid := response.results[0].oidIn
		if ptr.Oid != oid {
			err = errors.New(tr.Tr.Get("actual OID %s during smudge does not match expected %s", oid, ptr.Oid))
			return 0, errors.Wrap(err, tr.Tr.Get("smudge filter"))
		}

		for _, expected := range ptr.Extensions {
			actual := actualExts[expected.Name]
			if actual.name != expected.Name {
				err = errors.New(tr.Tr.Get("actual extension name '%s' does not match expected '%s'", actual.name, expected.Name))
				return 0, errors.Wrap(err, tr.Tr.Get("smudge filter"))
			}
			if actual.oidOut != expected.Oid {
				err = errors.New(tr.Tr.Get("actual OID %s for extension '%s' does not match expected %s", actual.oidOut, expected.Name, expected.Oid))
				return 0, errors.Wrap(err, tr.Tr.Get("smudge filter"))
			}
		}

		// setup reader
		reader, err = os.Open(response.file.Name())
		if err != nil {
			return 0, errors.Wrap(err, tr.Tr.Get("Error opening smudged file: %s", err))
		}
		defer reader.Close()
	}

	n, err := tools.CopyWithCallback(writer, reader, ptr.Size, cb)
	if err != nil {
		return n, errors.Wrap(err, tr.Tr.Get("Error reading from media file: %s", err))
	}

	return n, nil
}
