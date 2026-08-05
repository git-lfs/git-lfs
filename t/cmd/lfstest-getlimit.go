//go:build testtools
// +build testtools

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/pktline"
)

func main() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "--max-bufio-scan-token-size":
			fmt.Print(bufio.MaxScanTokenSize)
		case "--max-pktline-len":
			fmt.Print(pktline.MaxPacketLength)
		case "--max-pointer-size":
			fmt.Print(lfs.BlobSizeCutoff)
		case "--max-spool-mem-buffer-size":
			fmt.Print(tools.MemoryBufferLimit)
		default:
			fmt.Fprintf(os.Stderr, "unknown argument: %s\n", os.Args[1])
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "invalid arguments: %s\n", strings.Join(os.Args, " "))
		os.Exit(1)
	}

	os.Exit(0)
}
