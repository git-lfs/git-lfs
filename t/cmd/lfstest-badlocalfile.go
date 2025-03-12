//go:build testtools
// +build testtools

package main

import (
	"os"
	"strings"

	"github.com/git-lfs/pktline"
)

func main() {
	pl := pktline.NewPktline(os.Stdin, os.Stdout)

	// For any file Git asks us to clean, rename it and create a
	// 0-byte file in its place. This allows Git to continue to
	// read from the original file and send us its contents via
	// packets on stdin, while our filter-process command will
	// see the 0-byte file as soon as the first flush packet is
	// received.
	//
	// If we simply truncate the file to 0 bytes in situ, Git
	// may not be able to finish reading its full contents, and
	// might send us fewer packets than actually needed.

	var command, pathname string

	for {
		b, n, err := pl.ReadPacketWithLength()
		if err != nil {
			return
		}

		if n == 0 {
			pl.WriteFlush()
		} else if n == 1 {
			pl.WriteDelim()
		} else {
			parts := strings.SplitN(string(b), "=", 2)
			if len(parts) == 2 {
				if parts[0] == "command" {
					command = strings.TrimSuffix(parts[1], "\n")
				} else if parts[0] == "pathname" {
					pathname = strings.TrimSuffix(parts[1], "\n")
				}

				if command != "" && pathname != "" {
					if command == "clean" {
						if err = os.Rename(pathname, pathname + ".orig"); err != nil {
							panic(err)
						}

						f, err := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE, 0o644)
						if err != nil {
							panic(err)
						}
						if err = f.Truncate(0); err != nil {
							panic(err)
						}
						f.Close()
					}

					command = ""
					pathname = ""
				}
			}

			pl.WritePacket(b)
		}
	}
}
