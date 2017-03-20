package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
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

	skip := filterSmudgeSkip || cfg.Os.Bool("GIT_LFS_SKIP_SMUDGE", false)
	filter := filepathfilter.New(cfg.FetchIncludePaths(), cfg.FetchExcludePaths())

	var malformed []string

	for s.Scan() {
		var err error
		var w *git.PktlineWriter

		req := s.Request()

		s.WriteStatus(statusFromErr(nil))

		switch req.Header["command"] {
		case "clean":
			w = git.NewPktlineWriter(os.Stdout, cleanFilterBufferCapacity)
			err = clean(w, req.Payload, req.Header["pathname"])
		case "smudge":
			w = git.NewPktlineWriter(os.Stdout, smudgeFilterBufferCapacity)
			err = smudge(w, req.Payload, req.Header["pathname"], skip, filter)
		default:
			ExitWithError(fmt.Errorf("Unknown command %q", req.Header["command"]))
		}

		if errors.IsNotAPointerError(err) {
			malformed = append(malformed, req.Header["pathname"])
			err = nil
		}

		var status string
		if ferr := w.Flush(); ferr != nil {
			status = statusFromErr(ferr)
		} else {
			status = statusFromErr(err)
		}

		s.WriteStatus(status)
	}

	if len(malformed) > 0 {
		fmt.Fprintf(os.Stderr, "Encountered %d file(s) that should have been pointers, but weren't:\n", len(malformed))
		for _, m := range malformed {
			fmt.Fprintf(os.Stderr, "\t%s\n", m)
		}
	}

	if err := s.Err(); err != nil && err != io.EOF {
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
