package commands

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

var (
	pointerFile     string
	pointerCompare  string
	pointerStdin    bool
	pointerCheck    bool
	pointerStrict   bool
	pointerNoStrict bool
)

func pointerCommand(cmd *cobra.Command, args []string) {
	comparing := false
	something := false
	buildOid := ""
	compareOid := ""

	if pointerCheck {
		var r io.ReadCloser
		var err error

		if pointerStrict && pointerNoStrict {
			ExitWithError(errors.New(tr.Tr.Get("Cannot combine --strict with --no-strict")))
		}

		if len(pointerCompare) > 0 {
			ExitWithError(errors.New(tr.Tr.Get("Cannot combine --check with --compare")))
		}

		if len(pointerFile) > 0 {
			if pointerStdin {
				ExitWithError(errors.New(tr.Tr.Get("With --check, --file cannot be combined with --stdin")))
			}
			r, err = os.Open(pointerFile)
			if err != nil {
				ExitWithError(err)
			}
		} else if pointerStdin {
			r = io.NopCloser(os.Stdin)
		} else {
			ExitWithError(errors.New(tr.Tr.Get("Must specify either --file or --stdin with --compare")))
		}

		p, err := lfs.DecodePointer(r)
		if err != nil {
			os.Exit(1)
		}
		if pointerStrict && !p.Canonical {
			os.Exit(2)
		}
		r.Close()
		return
	}

	if len(pointerCompare) > 0 || pointerStdin {
		comparing = true
	}

	if len(pointerFile) > 0 {
		something = true
		buildFile, err := os.Open(pointerFile)
		if err != nil {
			Error(err.Error())
			os.Exit(1)
		}

		oidHash := sha256.New()
		size, err := io.Copy(oidHash, buildFile)
		buildFile.Close()

		if err != nil {
			Error(err.Error())
			os.Exit(1)
		}

		ptr := lfs.NewPointer(hex.EncodeToString(oidHash.Sum(nil)), size, nil)
		fmt.Fprint(os.Stderr, tr.Tr.Get("Git LFS pointer for %s", pointerFile), "\n\n")
		buf := &bytes.Buffer{}
		lfs.EncodePointer(io.MultiWriter(os.Stdout, buf), ptr)

		if comparing {
			buildOid, err = git.HashObject(bytes.NewReader(buf.Bytes()))
			if err != nil {
				Error(err.Error())
				os.Exit(1)
			}
			fmt.Fprint(os.Stderr, "\n", tr.Tr.Get("Git blob OID: %s", buildOid), "\n\n")
		}
	} else {
		comparing = false
	}

	if len(pointerCompare) > 0 || pointerStdin {
		something = true
		compFile, err := pointerReader()
		if err != nil {
			Error(err.Error())
			os.Exit(1)
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(compFile, buf)
		_, err = lfs.DecodePointer(tee)
		compFile.Close()

		pointerName := "STDIN"
		if !pointerStdin {
			pointerName = pointerCompare
		}
		fmt.Fprint(os.Stderr, tr.Tr.Get("Pointer from %s", pointerName), "\n\n")

		if err != nil {
			Error(err.Error())
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, buf.String())
		if comparing {
			compareOid, err = git.HashObject(bytes.NewReader(buf.Bytes()))
			if err != nil {
				Error(err.Error())
				os.Exit(1)
			}
			fmt.Fprint(os.Stderr, "\n", tr.Tr.Get("Git blob OID: %s", compareOid), "\n")
		}
	}

	if comparing && buildOid != compareOid {
		fmt.Fprint(os.Stderr, "\n", tr.Tr.Get("Pointers do not match"), "\n")
		os.Exit(1)
	}

	if !something {
		Error(tr.Tr.Get("Nothing to do!"))
		os.Exit(1)
	}
}

func pointerReader() (io.ReadCloser, error) {
	if len(pointerCompare) > 0 {
		if pointerStdin {
			return nil, errors.New(tr.Tr.Get("cannot read from STDIN and --pointer"))
		}

		return os.Open(pointerCompare)
	}

	requireStdin(tr.Tr.Get("The --stdin flag expects a pointer file from STDIN."))

	return os.Stdin, nil
}

func init() {
	RegisterCommand("pointer", pointerCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&pointerFile, "file", "f", "", "Path to a local file to generate the pointer from.")
		cmd.Flags().StringVarP(&pointerCompare, "pointer", "p", "", "Path to a local file containing a pointer built by another Git LFS implementation.")
		cmd.Flags().BoolVarP(&pointerStdin, "stdin", "", false, "Read a pointer built by another Git LFS implementation through STDIN.")
		cmd.Flags().BoolVarP(&pointerCheck, "check", "", false, "Check whether the given file is a Git LFS pointer.")
		cmd.Flags().BoolVarP(&pointerStrict, "strict", "", false, "Check whether the given Git LFS pointer is canonical.")
		cmd.Flags().BoolVarP(&pointerNoStrict, "no-strict", "", false, "Don't check whether the given Git LFS pointer is canonical.")
	})
}
