package commands

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/github/git-lfs/pointer"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
)

var (
	pointerBuild   string
	pointerCompare string
	pointerCmd     = &cobra.Command{
		Use:   "pointer",
		Short: "Build and compare pointers between different Git LFS implementations",
		Run:   pointerCommand,
	}
)

func pointerCommand(cmd *cobra.Command, args []string) {
	comparing := false
	something := false
	buildOid := ""
	compareOid := ""

	if len(pointerCompare) > 0 {
		comparing = true
	}

	if len(pointerBuild) > 0 {
		something = true
		buildFile, err := os.Open(pointerBuild)
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

		ptr := pointer.NewPointer(hex.EncodeToString(oidHash.Sum(nil)), size)
		fmt.Printf("Git LFS pointer for %s\n\n", pointerBuild)
		buf := &bytes.Buffer{}
		pointer.Encode(io.MultiWriter(os.Stdout, buf), ptr)

		if comparing {
			buildOid = gitHashObject(buf.Bytes())
			fmt.Printf("\nGit blob OID: %s\n\n", buildOid)
		}
	} else {
		comparing = false
	}

	if len(pointerCompare) > 0 {
		something = true
		compFile, err := os.Open(pointerCompare)
		if err != nil {
			Error(err.Error())
			os.Exit(1)
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(compFile, buf)
		_, err = pointer.Decode(tee)
		compFile.Close()

		fmt.Printf("Pointer from %s\n\n", pointerCompare)

		if err != nil {
			Error(err.Error())
			os.Exit(1)
		}

		fmt.Printf(buf.String())
		if comparing {
			compareOid = gitHashObject(buf.Bytes())
			fmt.Printf("\nGit blob OID: %s\n", compareOid)
		}
	}

	if comparing && buildOid != compareOid {
		fmt.Printf("\nPointers do not match\n")
		os.Exit(1)
	}

	if !something {
		Error("Nothing to do!")
		os.Exit(1)
	}
}

func gitHashObject(by []byte) string {
	cmd := exec.Command("git", "hash-object", "--stdin")
	cmd.Stdin = bytes.NewReader(by)
	out, err := cmd.Output()
	if err != nil {
		Error("Error building Git blob OID: %s", err)
		os.Exit(1)
	}

	return string(bytes.TrimSpace(out))
}

func init() {
	flags := pointerCmd.Flags()
	flags.StringVarP(&pointerBuild, "build", "b", "", "Path to a local file to generate the pointer from.")
	flags.StringVarP(&pointerCompare, "compare", "c", "", "Path to a local file containing a pointer built by another Git LFS implementation.")
	RootCmd.AddCommand(pointerCmd)
}
