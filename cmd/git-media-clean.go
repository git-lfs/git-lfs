package main

import (
	".."
	"../filters"
	"fmt"
	"os"
)

func main() {
	cleaned, err := gitmediafilters.Clean(os.Stdin)
	if err != nil {
		fmt.Println("Error cleaning asset")
		panic(err)
	}

	writer := cleaned.Writer(os.Stdout)
	defer writer.Close()
	gitmedia.Encode(writer, cleaned.Sha)
}
