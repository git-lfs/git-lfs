package main

import (
	"flag"
	"log"
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
		log.Fatalln("Unknown command:", *SubCommand)
	}
}
