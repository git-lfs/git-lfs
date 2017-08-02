package commands

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	pointerFile    string
	pointerCompare string
	pointerStdin   bool
)

func pointerCommand(cmd *cobra.Command, args []string) {
	comparing := false
	something := false
	buildOid := ""
	compareOid := ""

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
		fmt.Fprintf(os.Stderr, "Git LFS pointer for %s\n\n", pointerFile)
		buf := &bytes.Buffer{}
		lfs.EncodePointer(io.MultiWriter(os.Stdout, buf), ptr)

		if comparing {
			buildOid, err = git.HashObject(buf.Bytes())
			if err != nil {
				Error(err.Error())
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "\nGit blob OID: %s\n\n", buildOid)
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
		fmt.Fprintf(os.Stderr, "Pointer from %s\n\n", pointerName)

		if err != nil {
			Error(err.Error())
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, buf.String())
		if comparing {
			compareOid, err = git.HashObject(buf.Bytes())
			if err != nil {
				Error(err.Error())
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "\nGit blob OID: %s\n", compareOid)
		}
	}

	if comparing && buildOid != compareOid {
		fmt.Fprintf(os.Stderr, "\nPointers do not match\n")
		os.Exit(1)
	}

	if !something {
		Error("Nothing to do!")
		os.Exit(1)
	}
}

func pointerReader() (io.ReadCloser, error) {
	if len(pointerCompare) > 0 {
		if pointerStdin {
			return nil, errors.New("Cannot read from STDIN and --pointer.")
		}

		return os.Open(pointerCompare)
	}

	requireStdin("The --stdin flag expects a pointer file from STDIN.")

	return os.Stdin, nil
}

func init() {
	RegisterCommand("pointer", pointerCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&pointerFile, "file", "f", "", "Path to a local file to generate the pointer from.")
		cmd.Flags().StringVarP(&pointerCompare, "pointer", "p", "", "Path to a local file containing a pointer built by another Git LFS implementation.")
		cmd.Flags().BoolVarP(&pointerStdin, "stdin", "", false, "Read a pointer built by another Git LFS implementation through STDIN.")
	})
}
