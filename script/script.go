package main

import (
	"flag"
	"log"
)

type Release struct {
	Label    string
	Filename string
	SHA256   string
}

var SubCommand = flag.String("cmd", "", "Command: build or release")

func main() {
	flag.Parse()
	switch *SubCommand {
	case "build":
		mainBuild()
	case "release":
		mainRelease()
	case "integration":
		mainIntegration()
	default:
		log.Fatalln("Unknown command:", *SubCommand)
	}
}
