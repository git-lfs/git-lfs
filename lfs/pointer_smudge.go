package lfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/vendor/_nuts/github.com/cheggaaa/pb"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
	contentaddressable "github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/go-contentaddressable"
)

var (
	DownloadDeclinedError = errors.New("File missing and download is not allowed")
)

func PointerSmudgeToFile(filename string, ptr *Pointer, download bool, cb CopyCallback) error {
	os.MkdirAll(filepath.Dir(filename), 0755)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Could not create working directory file: %v", err)
	}
	defer file.Close()
	if err := PointerSmudge(file, ptr, filename, download, cb); err != nil {
		if err == DownloadDeclinedError {
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
func PointerSmudge(writer io.Writer, ptr *Pointer, workingfile string, download bool, cb CopyCallback) error {
	mediafile, err := LocalMediaPath(ptr.Oid)
	if err != nil {
		return err
	}

	stat, statErr := os.Stat(mediafile)
	if statErr == nil && stat != nil {
		fileSize := stat.Size()
		if fileSize == 0 || fileSize != ptr.Size {
			tracerx.Printf("Removing %s, size %d is invalid", mediafile, fileSize)
			os.RemoveAll(mediafile)
			stat = nil
		}
	}

	var wErr *WrappedError
	if statErr != nil || stat == nil {
		if download {
			wErr = downloadFile(writer, ptr, workingfile, mediafile, cb)
		} else {
			return DownloadDeclinedError
		}
	} else {
		wErr = readLocalFile(writer, ptr, mediafile, workingfile, cb)
	}

	if wErr != nil {
		return &SmudgeError{ptr.Oid, mediafile, wErr}
	}

	return nil
}

// PointerSmudgeObject uses a Pointer and objectResource to download the object to the
// media directory. It does not write the file to the working directory.
func PointerSmudgeObject(ptr *Pointer, obj *objectResource, cb CopyCallback) error {
	mediafile, err := LocalMediaPath(obj.Oid)
	if err != nil {
		return err
	}

	stat, statErr := os.Stat(mediafile)
	if statErr == nil && stat != nil {
		fileSize := stat.Size()
		if fileSize == 0 || fileSize != obj.Size {
			tracerx.Printf("Removing %s, size %d is invalid", mediafile, fileSize)
			os.RemoveAll(mediafile)
			stat = nil
		}
	}

	if statErr != nil || stat == nil {
		wErr := downloadObject(ptr, obj, mediafile, cb)

		if wErr != nil {
			return &SmudgeError{obj.Oid, mediafile, wErr}
		}
	}

	return nil
}

func downloadObject(ptr *Pointer, obj *objectResource, mediafile string, cb CopyCallback) *WrappedError {
	reader, size, wErr := DownloadObject(obj)
	if reader != nil {
		defer reader.Close()
	}

	// TODO this can be unified with the same code in downloadFile
	if wErr != nil {
		wErr.Errorf("Error downloading %s.", mediafile)
		return wErr
	}

	if ptr.Size == 0 {
		ptr.Size = size
	}

	mediaFile, err := contentaddressable.NewFile(mediafile)
	if err != nil {
		return Errorf(err, "Error opening media file buffer.")
	}

	_, err = CopyWithCallback(mediaFile, reader, ptr.Size, cb)
	if err == nil {
		err = mediaFile.Accept()
	}
	mediaFile.Close()

	if err != nil {
		return Errorf(err, "Error buffering media file.")
	}

	return nil
}

func downloadFile(writer io.Writer, ptr *Pointer, workingfile, mediafile string, cb CopyCallback) *WrappedError {
	fmt.Fprintf(os.Stderr, "Downloading %s (%s)\n", workingfile, pb.FormatBytes(ptr.Size))
	reader, size, wErr := Download(filepath.Base(mediafile))
	if reader != nil {
		defer reader.Close()
	}

	if wErr != nil {
		wErr.Errorf("Error downloading %s.", mediafile)
		return wErr
	}

	if ptr.Size == 0 {
		ptr.Size = size
	}

	mediaFile, err := contentaddressable.NewFile(mediafile)
	if err != nil {
		return Errorf(err, "Error opening media file buffer.")
	}

	_, err = CopyWithCallback(mediaFile, reader, ptr.Size, cb)
	if err == nil {
		err = mediaFile.Accept()
	}
	mediaFile.Close()

	if err != nil {
		return Errorf(err, "Error buffering media file.")
	}

	return readLocalFile(writer, ptr, mediafile, workingfile, nil)
}

func readLocalFile(writer io.Writer, ptr *Pointer, mediafile string, workingfile string, cb CopyCallback) *WrappedError {
	reader, err := os.Open(mediafile)
	if err != nil {
		return Errorf(err, "Error opening media file.")
	}
	defer reader.Close()

	if ptr.Size == 0 {
		if stat, _ := os.Stat(mediafile); stat != nil {
			ptr.Size = stat.Size()
		}
	}

	if len(ptr.Extensions) > 0 {
		registeredExts := Config.Extensions()
		extensions := make(map[string]Extension)
		for _, ptrExt := range ptr.Extensions {
			ext, ok := registeredExts[ptrExt.Name]
			if !ok {
				err := fmt.Errorf("Extension '%s' is not configured.", ptrExt.Name)
				return Error(err)
			}
			ext.Priority = ptrExt.Priority
			extensions[ext.Name] = ext
		}
		exts, err := SortExtensions(extensions)
		if err != nil {
			return Error(err)
		}

		// pipe extensions in reverse order
		var extsR []Extension
		for i, _ := range exts {
			ext := exts[len(exts)-1-i]
			extsR = append(extsR, ext)
		}

		request := &pipeRequest{"smudge", reader, workingfile, extsR}

		response, err := pipeExtensions(request)
		if err != nil {
			return Error(err)
		}

		actualExts := make(map[string]*pipeExtResult)
		for _, result := range response.results {
			actualExts[result.name] = result
		}

		// verify name, order, and oids
		oid := response.results[0].oidIn
		if ptr.Oid != oid {
			err = fmt.Errorf("Actual oid %s during smudge does not match expected %s", oid, ptr.Oid)
			return Error(err)
		}

		for _, expected := range ptr.Extensions {
			actual := actualExts[expected.Name]
			if actual.name != expected.Name {
				err = fmt.Errorf("Actual extension name '%s' does not match expected '%s'", actual.name, expected.Name)
				return Error(err)
			}
			if actual.oidOut != expected.Oid {
				err = fmt.Errorf("Actual oid %s for extension '%s' does not match expected %s", actual.oidOut, expected.Name, expected.Oid)
				return Error(err)
			}
		}

		// setup reader
		reader, err = os.Open(response.file.Name())
		if err != nil {
			return Errorf(err, "Error opening smudged file.")
		}
		defer reader.Close()
	}

	_, err = CopyWithCallback(writer, reader, ptr.Size, cb)
	if err != nil {
		return Errorf(err, "Error reading from media file.")
	}

	return nil
}

type SmudgeError struct {
	Oid      string
	Filename string
	*WrappedError
}
