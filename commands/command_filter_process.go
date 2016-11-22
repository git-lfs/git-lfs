package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
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

// filterSmudgeSkip is a command-line flag owned by the `filter-process` command
// dictating whether or not to skip the smudging process, leaving pointers as-is
// in the working tree.
var filterSmudgeSkip bool

// filterSmudge is a gateway to the `smudge()` function and serves to bail out
// immediately if the pointer decoded from "from" has no data (i.e., is empty).
// This function, unlike the implementation found in the legacy smudge command,
// only combines the `io.Reader`s when necessary, since the implementation
// found in `*git.PacketReader` blocks while waiting for the following packet.
func filterSmudge(to io.Writer, from io.Reader, filename string) error {
	var pbuf bytes.Buffer
	from = io.TeeReader(from, &pbuf)

	ptr, err := lfs.DecodePointer(from)
	if err != nil {
		// If we tried to decode a pointer out of the data given to us,
		// and the file was _empty_, write out an empty file in
		// response. This occurs because when the clean filter
		// encounters an empty file, and writes out an empty file,
		// instead of a pointer.
		if pbuf.Len() == 0 {
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
	var err error

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
	for s.Scan() && err == nil {
		var w *git.PktlineWriter

		req := s.Request()

		s.WriteStatus(statusFromErr(nil))

		switch req.Header["command"] {
		case "clean":
			w = git.NewPktlineWriter(os.Stdout, cleanFilterBufferCapacity)
			err = clean(w, req.Payload, req.Header["pathname"])
		case "smudge":
			w = git.NewPktlineWriter(os.Stdout, smudgeFilterBufferCapacity)
			err = filterSmudge(w, req.Payload, req.Header["pathname"])
		default:
			fmt.Errorf("Unknown command %s", cmd)
			break Scan
		}

		if err != nil {
			s.WriteStatus(statusFromErr(w.Flush()))
		}
	}

	if err == nil {
		err = s.Err()
	}

	if err != nil {
		ExitWithError(err)
	}
}

// statusFromErr returns the status code that should be sent over the filter
// protocol based on a given error, "err".
func statusFromErr(err error) string {
	if err != nil && err != io.EOF {
		return "error"
	}
	return "success"
}

func init() {
	RegisterCommand("filter-process", filterCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&filterSmudgeSkip, "skip", "s", false, "")
	})
}
