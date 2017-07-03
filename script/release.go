package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

var (
	ReleaseId    = flag.Int("id", 0, "git-lfs/git-lfs Release ID")
	uploadUrlFmt = "https://uploads.github.com/repos/git-lfs/git-lfs/releases/%d/assets?%s"
)

func mainRelease() {
	if *ReleaseId < 1 {
		log.Println("Need a valid git-lfs/git-lfs release id.")
		log.Fatalln("usage: script/release -id")
	}

	file, err := os.Open("bin/releases/build_matrix.json")
	if err != nil {
		log.Println("Error opening build_matrix.json:", err)
		log.Fatalln("Ensure `script/bootstrap -all` has completed successfully")
	}

	defer file.Close()

	buildMatrix := make(map[string]Release)
	if err := json.NewDecoder(file).Decode(&buildMatrix); err != nil {
		log.Fatalln("Error reading build_matrix.json:", err)
	}

	for _, rel := range buildMatrix {
		release(rel)
		fmt.Println()
	}

	fmt.Println("SHA-256 hashes:")
	for _, rel := range buildMatrix {
		fmt.Printf("**%s**\n%s\n\n", rel.Filename, rel.SHA256)
	}
}

func release(rel Release) {
	query := url.Values{}
	query.Add("name", rel.Filename)
	query.Add("label", rel.Label)

	args := []string{
		"-in",
		"-H", "Content-Type: application/octet-stream",
		"-X", "POST",
		"--data-binary", "@bin/releases/" + rel.Filename,
		fmt.Sprintf(uploadUrlFmt, *ReleaseId, query.Encode()),
	}

	fmt.Println("curl", strings.Join(args, " "))

	cmd := exec.Command("curl", args...)

	by, err := cmd.Output()
	if err != nil {
		log.Fatalln("Error running curl:", err)
	}

	fmt.Println(string(by))
}
