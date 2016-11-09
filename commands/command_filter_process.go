package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

const (
	// cleanFilterBufferCapacity is the desired capacity of the
	// `*git.PacketWriter`'s internal buffer when the filter protocol
	// dictates the "clean" command. 512 bytes is (in most cases) enough to
	// hold an entire LFS pointer in memory.
	cleanFilterBufferCapacity = 512

	// smudgeFilterBufferCapacity is the desired capacity of the
	// `*git.PacketWriter`'s internal buffer when the filter protocol
	// dictates the "smudge" command.
	smudgeFilterBufferCapacity = git.MaxPacketLength
)

var (
	filterSmudgeSkip = false
)

func filterSmudge(from io.Reader, to io.Writer, filename string) error {
	var pbuf bytes.Buffer
	from = io.TeeReader(from, &pbuf)

	ptr, err := lfs.DecodePointer(from)
	if err != nil {
		// If we tried to decode a pointer out of the data given to us,
		// and the file was _empty_, write out an empty file in
		// response. This occurs because when the clean filter
		// encounters an empty file, and writes out an empty file,
		// instead of a pointer.
		//
		// TODO(taylor): figure out if there is more data on the reader,
		// and buffer that as well.
		if len(pbuf.Bytes()) == 0 {
			if _, cerr := io.Copy(to, &pbuf); cerr != nil {
				Panic(cerr, "Error writing data to stdout:")
			}
			return nil
		}

		return err
	}

	lfs.LinkOrCopyFromReference(ptr.Oid, ptr.Size)

	return smudge(to, ptr, filename, filterSmudgeSkip)
}

func filterCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git filter process")
	lfs.InstallHooks(false)

	s := git.NewFilterProcessScanner(os.Stdin, os.Stdout)

	if err := s.Init(); err != nil {
		ExitWithError(err)
	}
	if err := s.NegotiateCapabilities(); err != nil {
		ExitWithError(err)
	}

Scan:
	for s.Scan() {
		var err error
		var w *git.PacketWriter

		req := s.Request()

		s.WriteStatus("success")

		switch req.Header["command"] {
		case "clean":
			w = git.NewPacketWriter(os.Stdout, cleanFilterBufferCapacity)
			err = clean(req.Payload, w, req.Header["pathname"])
		case "smudge":
			w = git.NewPacketWriter(os.Stdout, smudgeFilterBufferCapacity)
			err = filterSmudge(req.Payload, w, req.Header["pathname"])
		default:
			fmt.Errorf("Unknown command %s", cmd)
			break Scan
		}

		var status string
		if ferr := w.Flush(); ferr != nil {
			status = "error"
		} else {
			if err != nil && err != io.EOF {
				status = "error"
			} else {
				status = "success"
			}
		}
		s.WriteStatus(status)
	}

	if err := s.Err(); err != nil && err != io.EOF {
		ExitWithError(err)
	}
}

func init() {
	RegisterCommand("filter-process", filterCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&filterSmudgeSkip, "skip", "s", false, "")
	})
}
