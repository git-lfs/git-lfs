package main

import (
	"flag"
	"fmt"
	"os"
)

type Release struct {
	Label    string
	Filename string
}

var SubCommand = flag.String("cmd", "", "Command: build or release")

func main() {
	flag.Parse()
	switch *SubCommand {
	case "build":
		mainBuild()
	case "release":
		mainRelease()
	default:
		fmt.Println("Unknown command:", *SubCommand)
		os.Exit(1)
	}
}
