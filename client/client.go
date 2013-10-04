package gitmediaclient

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func Send(filename string) error {
	sha := filepath.Base(filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", "http://localhost:8080/objects/"+sha, file)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	fmt.Printf("Sending %s from %s: %d\n", sha, filename, res.StatusCode)
	return nil
}
