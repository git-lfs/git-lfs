package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errors"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/progress"
	"github.com/spf13/cobra"
)

var (
	filterSmudgeSkip = false
)

func clean(reader io.Reader, fileName string) ([]byte, error) {
	var cb progress.CopyCallback
	var file *os.File
	var fileSize int64
	if len(fileName) > 0 {
		stat, err := os.Stat(fileName)
		if err == nil && stat != nil {
			fileSize = stat.Size()

			localCb, localFile, err := lfs.CopyCallbackFile("clean", fileName, 1, 1)
			if err != nil {
				Error(err.Error())
			} else {
				cb = localCb
				file = localFile
			}
		}
	}

	cleaned, err := lfs.PointerClean(reader, fileName, fileSize, cb)
	if file != nil {
		file.Close()
	}

	if cleaned != nil {
		defer cleaned.Teardown()
	}

	if errors.IsCleanPointerError(err) {
		// TODO: report errors differently!
		// os.Stdout.Write(errors.GetContext(err, "bytes").([]byte))
		return errors.GetContext(err, "bytes").([]byte), nil
	}

	if err != nil {
		Panic(err, "Error cleaning asset.")
	}

	tmpfile := cleaned.Filename
	mediafile, err := lfs.LocalMediaPath(cleaned.Oid)
	if err != nil {
		Panic(err, "Unable to get local media path.")
	}

	if stat, _ := os.Stat(mediafile); stat != nil {
		if stat.Size() != cleaned.Size && len(cleaned.Pointer.Extensions) == 0 {
			Exit("Files don't match:\n%s\n%s", mediafile, tmpfile)
		}
		Debug("%s exists", mediafile)
	} else {
		if err := os.Rename(tmpfile, mediafile); err != nil {
			Panic(err, "Unable to move %s to %s\n", tmpfile, mediafile)
		}

		Debug("Writing %s", mediafile)
	}

	return []byte(cleaned.Pointer.Encoded()), nil
}

func smudge(reader io.Reader, filename string) ([]byte, error) {
	ptr, err := lfs.DecodePointer(reader)
	if err != nil {
		// mr := io.MultiReader(b, reader)
		// _, err := io.Copy(os.Stdout, mr)
		// if err != nil {
		// 	Panic(err, "Error writing data to stdout:")
		// }
		var content []byte
		reader.Read(content)
		return content, nil
	}

	lfs.LinkOrCopyFromReference(ptr.Oid, ptr.Size)

	cb, file, err := lfs.CopyCallbackFile("smudge", filename, 1, 1)
	if err != nil {
		Error(err.Error())
	}

	cfg := config.Config
	download := lfs.FilenamePassesIncludeExcludeFilter(filename, cfg.FetchIncludePaths(), cfg.FetchExcludePaths())

	if filterSmudgeSkip || cfg.Os.Bool("GIT_LFS_SKIP_SMUDGE", false) {
		download = false
	}

	buf := new(bytes.Buffer)
	err = ptr.Smudge(buf, filename, download, TransferManifest(), cb)
	if file != nil {
		file.Close()
	}

	if err != nil {
		// Download declined error is ok to skip if we weren't requesting download
		if !(errors.IsDownloadDeclinedError(err) && !download) {
			LoggedError(err, "Error downloading object: %s (%s)", filename, ptr.Oid)
			if !cfg.SkipDownloadErrors() {
				// TODO: What to do best here?
				os.Exit(2)
			}
		}

		return []byte(ptr.Encoded()), nil
	}

	return buf.Bytes(), nil
}

func filterCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git filter process")
	lfs.InstallHooks(false)

	s := git.NewObjectScanner(os.Stdin, os.Stdout)

	if err := s.Init(); err != nil {
		ExitWithError(err)
	}
	if err := s.NegotiateCapabilities(); err != nil {
		ExitWithError(err)
	}

	for {
		req, err := s.ReadRequest()
		if err != nil {
			break
		}

		// TODO:
		// ReadRequest should return data as Reader instead of []byte ?!
		// clean/smudge should also take a Writer instead of returning []byte
		var outputData []byte
		switch req.Header["command"] {
		case "clean":
			outputData, _ = clean(bytes.NewReader(req.Payload), req.Header["pathname"])
		case "smudge":
			outputData, _ = smudge(bytes.NewReader(req.Payload), req.Header["pathname"])
		default:
			fmt.Errorf("Unknown command %s", cmd)
			break
		}

		s.WriteResponse(outputData)
	}
}

func init() {
	RegisterCommand("filter", filterCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&filterSmudgeSkip, "skip", "s", false, "")
	})
}
