package main

import (
	".."
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

func main() {
	buf := make([]byte, 10)
	os.Stdin.Read(buf)
	slices := bytes.SplitN(buf, []byte("\n"), 2)
	headerlen, err := strconv.Atoi(string(slices[0])[2:])
	if err != nil {
		fmt.Printf("Error reading header:\n%s\n", string(buf))
		panic(err)
	}

	buf = make([]byte, headerlen-len(slices[1]))
	os.Stdin.Read(buf)

	meta := &gitmedia.LargeAsset{}
	dec := json.NewDecoder(os.Stdin)
	dec.Decode(meta)

	mediafile := gitmedia.LocalMediaPath(meta.SHA1)
	file, err := os.Open(mediafile)
	if err != nil {
		fmt.Printf("Error reading file from local media dir: %s\n", mediafile)
		panic(err)
	}
	defer file.Close()

	io.Copy(os.Stdout, file)
}
