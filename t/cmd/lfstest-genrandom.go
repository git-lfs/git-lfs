//go:build testtools
// +build testtools

package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
)

const usageFmt = "Usage: %s [--base64|--base64url] [<size>]\n"

func main() {
	offset := 1
	b64 := false
	b64url := false
	if len(os.Args) > offset && (os.Args[offset] == "--base64" || os.Args[offset] == "--base64url") {
		b64 = true
		b64url = os.Args[offset] == "--base64url"
		offset += 1
	}

	if len(os.Args) > offset+1 {
		fmt.Fprintf(os.Stderr, usageFmt, os.Args[0])
		os.Exit(2)
	}

	var count uint64 = ^uint64(0)
	if len(os.Args) == offset+1 {
		var err error
		if count, err = strconv.ParseUint(os.Args[offset], 10, 64); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading size: %s, %v\n", os.Args[offset], err)
			os.Exit(3)
		}
	}

	b := make([]byte, 32)
	bb := make([]byte, max(base64.RawStdEncoding.EncodedLen(len(b)), base64.RawURLEncoding.EncodedLen(len(b))))
	for count > 0 {
		n, err := rand.Read(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading random bytes: %v\n", err)
			os.Exit(4)
		}
		if b64 {
			if b64url {
				base64.RawURLEncoding.Encode(bb, b[:n])
				n = base64.RawURLEncoding.EncodedLen(n)
			} else {
				base64.RawStdEncoding.Encode(bb, b[:n])
				n = base64.RawStdEncoding.EncodedLen(n)
			}
		}

		num := min(uint64(n), count)
		if b64 {
			err = binary.Write(os.Stdout, binary.LittleEndian, bb[:num])
		} else {
			err = binary.Write(os.Stdout, binary.LittleEndian, b[:num])
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing random bytes: %v\n", err)
			os.Exit(5)
		}
		count -= num
	}
}
