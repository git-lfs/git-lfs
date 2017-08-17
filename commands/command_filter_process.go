package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tq"
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

	caps, err := s.NegotiateCapabilities()
	if err != nil {
		ExitWithError(err)
	}

	var supportsDelay bool
	for _, cap := range caps {
		if cap == "capability=delay" {
			supportsDelay = true
			break
		}
	}

	skip := filterSmudgeSkip || cfg.Os.Bool("GIT_LFS_SKIP_SMUDGE", false)
	filter := filepathfilter.New(cfg.FetchIncludePaths(), cfg.FetchExcludePaths())

	ptrs := make(map[string]*lfs.Pointer)

	var q *tq.TransferQueue
	closeOnce := new(sync.Once)
	available := make(chan *tq.Transfer)

	if supportsDelay {
		q = tq.NewTransferQueue(tq.Download, getTransferManifest(), cfg.CurrentRemote)
		go infiniteTransferBuffer(q, available)
	}

	var malformed []string
	var malformedOnWindows []string
	for s.Scan() {
		var n int64
		var err error
		var delayed bool
		var w *git.PktlineWriter

		req := s.Request()

		if !(req.Header["command"] == "smudge" && req.Header["can-delay"] == "1") && !(req.Header["command"] == "list_available_blobs") {
			s.WriteStatus(statusFromErr(nil))
		}

		switch req.Header["command"] {
		case "clean":
			w = git.NewPktlineWriter(os.Stdout, cleanFilterBufferCapacity)

			var ptr *lfs.Pointer
			ptr, err = clean(w, req.Payload, req.Header["pathname"], -1)

			if ptr != nil {
				n = ptr.Size
			}
		case "smudge":
			w = git.NewPktlineWriter(os.Stdout, smudgeFilterBufferCapacity)
			if req.Header["can-delay"] == "1" {
				var ptr *lfs.Pointer

				n, delayed, ptr, err = delayedSmudge(s, w, req.Payload, q, req.Header["pathname"], skip, filter)

				if delayed {
					ptrs[req.Header["pathname"]] = ptr
				}
			} else {
				from, ferr := incomingOrCached(req.Payload, ptrs[req.Header["pathname"]])
				if ferr != nil {
					break
				}

				n, err = smudge(w, from, req.Header["pathname"], skip, filter)
				if err == nil {
					delete(ptrs, req.Header["pathname"])
				}
			}
		case "list_available_blobs":
			closeOnce.Do(func() {
				// The first time that Git sends us the
				// 'list_available_blobs' command, it is given
				// that no more smudge commands will be issued
				// with _new_ checkout entries.
				//
				// This means that, by the time that we're here,
				// we have seen all entries in the checkout, and
				// should therefore instruct the transfer queue
				// to make a batch out of whatever remaining
				// items it has, and then close itself.
				//
				// This function call is wrapped in a
				// `sync.(*Once).Do()` call so we only call
				// `q.Wait()` once, and is called via a
				// goroutine since `q.Wait()` is blocking.
				go q.Wait()
			})

			// The first, and all subsequent calls to
			// list_available_blobs, we read items from `tq.Watch()`
			// until a read from that channel becomes blocking (in
			// other words, we read until there are no more items
			// immediately ready to be sent back to Git).
			paths := pathnames(readAvailable(available))
			if len(paths) == 0 {
				// If `len(paths) == 0`, `tq.Watch()` has
				// closed, indicating that all items have been
				// completely processed, and therefore, sent
				// back to Git for checkout.
				for path, _ := range ptrs {
					// If we sent a path to Git but it
					// didn't ask for the smudge contents,
					// that path is available and Git should
					// accept it later.
					paths = append(paths, fmt.Sprintf("pathname=%s", path))
				}
			}
			err = s.WriteList(paths)
		default:
			ExitWithError(fmt.Errorf("Unknown command %q", req.Header["command"]))
		}

		if errors.IsNotAPointerError(err) {
			malformed = append(malformed, req.Header["pathname"])
			err = nil
		} else if possiblyMalformedObjectSize(n) {
			malformedOnWindows = append(malformedOnWindows, req.Header["pathname"])
		}

		if delayed {
			// If we delayed, there is no need to write a flush
			// packet since no content was written.
			w = nil
		}

		var status string
		if ferr := w.Flush(); ferr != nil {
			status = statusFromErr(ferr)
		} else {
			if delayed {
				// If the flush operation succeeded, write that
				// we were delayed, or encountered an error.
				// the checkout operation.
				status = delayedStatusFromErr(err)
			} else {
				// If we responded with content, report the
				// status of that operation instead.
				status = statusFromErr(err)
			}
		}

		s.WriteStatus(status)
	}

	if len(malformed) > 0 {
		fmt.Fprintf(os.Stderr, "Encountered %d file(s) that should have been pointers, but weren't:\n", len(malformed))
		for _, m := range malformed {
			fmt.Fprintf(os.Stderr, "\t%s\n", m)
		}
	}

	if len(malformedOnWindows) > 0 {
		fmt.Fprintf(os.Stderr, "Encountered %d file(s) that may not have been copied correctly on Windows:\n")

		for _, m := range malformedOnWindows {
			fmt.Fprintf(os.Stderr, "\t%s\n", m)
		}

		fmt.Fprintf(os.Stderr, "\nSee: `git lfs help smudge` for more details.\n")
	}

	if err := s.Err(); err != nil && err != io.EOF {
		ExitWithError(err)
	}
}

// infiniteTransferBuffer streams the results of q.Watch() into "available" as
// if available had an infinite channel buffer.
func infiniteTransferBuffer(q *tq.TransferQueue, available chan<- *tq.Transfer) {
	// Stream results from q.Watch() into chan "available" via an infinite
	// buffer.

	watch := q.Watch()

	// pending is used to keep track of an ordered list of available
	// `*tq.Transfer`'s that cannot be written to "available" without
	// blocking.
	var pending []*tq.Transfer

	for {
		if len(pending) > 0 {
			select {
			case t, ok := <-watch:
				if !ok {
					// If the list of pending elements is
					// non-empty, stream them out (even if
					// they block), and then close().
					for _, t = range pending {
						available <- t
					}
					close(available)
					return
				}
				pending = append(pending, t)
			case available <- pending[0]:
				// Otherwise, dequeue and shift the first
				// element from pending onto available.
				pending = pending[1:]
			}
		} else {
			t, ok := <-watch
			if !ok {
				// If watch is closed, the "tq" is done, and
				// there are no items on the buffer.  Return
				// immediately.
				close(available)
				return
			}

			select {
			case available <- t:
			// Copy an item directly from <-watch onto available<-.
			default:
				// Otherwise, if that would have blocked, make
				// the new read pending.
				pending = append(pending, t)
			}
		}
	}
}

// incomingOrCached returns an io.Reader that is either the contents of the
// given io.Reader "r", or the encoded contents of "ptr". It returns an error if
// there was an error reading from "r".
//
// This is done because when a `command=smudge` with `can-delay=0` is issued,
// the entry's contents are not sent, and must be re-encoded from the stored
// pointer corresponding to the request's filepath.
func incomingOrCached(r io.Reader, ptr *lfs.Pointer) (io.Reader, error) {
	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, r, 1024); err != nil {
		if err != io.EOF {
			return nil, err
		}
	}

	if buf.Len() < 1024 && ptr != nil {
		return strings.NewReader(ptr.Encoded()), nil
	}
	return io.MultiReader(&buf, r), nil
}

// readAvailable satisfies the accumulation semantics for the
// 'list_available_blobs' command. It accumulates items until:
//
// 1. Reading from the channel of available items blocks, or ...
// 2. There is one item available, or ...
// 3. The 'tq.TransferQueue' is completed.
func readAvailable(ch <-chan *tq.Transfer) []*tq.Transfer {
	ts := make([]*tq.Transfer, 0, 100)

	for {
		select {
		case t, ok := <-ch:
			if !ok {
				return ts
			}
			ts = append(ts, t)
		default:
			if len(ts) > 0 {
				return ts
			}

			t, ok := <-ch
			if !ok {
				return ts
			}
			return append(ts, t)
		}
	}

	return ts
}

// pathnames formats a list of *tq.Transfers as a valid response to the
// 'list_available_blobs' command.
func pathnames(ts []*tq.Transfer) []string {
	pathnames := make([]string, 0, len(ts))
	for _, t := range ts {
		pathnames = append(pathnames, fmt.Sprintf("pathname=%s", t.Name))
	}

	return pathnames
}

// statusFromErr returns the status code that should be sent over the filter
// protocol based on a given error, "err".
func statusFromErr(err error) string {
	if err != nil && err != io.EOF {
		return "error"
	}
	return "success"
}

// delayedStatusFromErr returns the status code that should be sent over the
// filter protocol based on a given error, "err" when the blob smudge operation
// was delayed.
func delayedStatusFromErr(err error) string {
	if err != nil && err != io.EOF {
		return "error"
	}
	return "delayed"
}

func init() {
	RegisterCommand("filter-process", filterCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&filterSmudgeSkip, "skip", "s", false, "")
	})
}
