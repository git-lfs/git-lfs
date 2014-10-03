package commands

import (
	"bytes"
	"fmt"
	"github.com/github/git-media/git"
	"github.com/github/git-media/pointer"
	"github.com/spf13/cobra"
	"strconv"
)

var (
	scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Scan for git media files",
		Run:   scanCommand,
	}
)

func scanCommand(cmd *cobra.Command, args []string) {

	// Gets all objects git knows about
	var buf bytes.Buffer
	objects, _ := git.RevListObjects("", "", true)
	for _, o := range objects {
		buf.WriteString(o.Sha1 + "\n")
	}

	// Get type and size info for all objects
	objects, _ = git.CatFileBatchCheck(&buf)

	// Pull out git objects that are type blob and size < 200 bytes.
	// These are the likely git media pointer files
	var mediaObjects bytes.Buffer
	for _, o := range objects {
		if o.Type == "blob" && o.Size < 200 {
			mediaObjects.WriteString(o.Sha1 + "\n")
		}
	}

	// Take all of the git media shas and pull out the pointer file contents
	// It comes out of here in the format:
	// <sha1> <type> <size><LF>
	// <contents><LF>
	// This string contains all the data, so we parse it out below
	data, _ := git.CatFileBatch(&mediaObjects)

	r := bytes.NewBufferString(data)

	pointers := make([]*pointer.Pointer, 0)
	for {
		l, err := r.ReadBytes('\n')
		if err != nil {
			break
		}

		s, _ := strconv.Atoi(string(bytes.Fields(l)[2]))

		nbuf := make([]byte, s)
		_, err = r.Read(nbuf)
		if err != nil {
			fmt.Println("Shouldn't have this :/")
			break
		}

		p, err := pointer.Decode(bytes.NewBuffer(nbuf))
		if err == nil {
			pointers = append(pointers, p)
		}

		_, err = r.ReadBytes('\n') // Extra \n inserted by cat-file
		if err != nil {
			break
		}
	}

	// Now we have Pointer objects for all git media files.
	// What can we do with them?
	// Create link files
	// Offer to download
	// ?
	// Profit

	for _, p := range pointers {
		fmt.Println("Git Media OID:", p.Oid)
	}
}

func init() {
	RootCmd.AddCommand(scanCmd)
}
