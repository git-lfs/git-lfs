package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
)

type Release struct {
	Label    string
	Filename string
}

var (
	ReleaseId    = flag.Int("id", 0, "github/git-media Release ID")
	uploadUrlFmt = "https://uploads.github.com/repos/github/git-media/releases/%d/assets?%s"
)

func main() {
	flag.Parse()

	if *ReleaseId < 1 {
		fmt.Println("Need a valid github/git-media release id.")
		fmt.Println("usage: script/release -id")
		os.Exit(1)
	}

	file, err := os.Open("bin/releases/build_matrix.json")
	if err != nil {
		fmt.Println("Error opening build_matrix.json:", err)
		fmt.Println("Ensure `script/bootstrap -all` has completed successfully")
		os.Exit(1)
	}

	defer file.Close()

	buildMatrix := make(map[string]Release)
	if err := json.NewDecoder(file).Decode(&buildMatrix); err != nil {
		fmt.Println("Error reading build_matrix.json:", err)
		os.Exit(1)
	}

	for _, rel := range buildMatrix {
		release(rel)
	}
}

func release(rel Release) {
	query := url.Values{}
	query.Add("name", rel.Filename)
	query.Add("label", rel.Label)

	cmd := exec.Command("curl", "-in",
		"-H", "Content-Type: application/octet-stream",
		"-X", "POST",
		"-d", "@bin/releases/"+rel.Filename,
		fmt.Sprintf(uploadUrlFmt, *ReleaseId, query.Encode()))

	by, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running curl:", err)
		os.Exit(1)
	}

	fmt.Println(string(by))
}
