package pointer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/github/git-media/git"
	"github.com/github/git-media/gitmedia"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	linkTemplate = `oid %s
name %s
`
)

type Link struct {
	Oid  string
	Name string
}

func (p *Pointer) CreateLink(filename string) error {
	hash, err := git.NewHashObject([]byte(p.Encoded()))
	if err != nil {
		return err
	}

	linkFile, err := gitmedia.LocalLinkPath(hash)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(linkFile, []byte(fmt.Sprintf(linkTemplate, p.Oid, filename)), 0644)
}

func FindLink(sha1 string) (*Link, error) {
	linkPath := filepath.Join(gitmedia.LocalLinkDir, sha1[0:2], sha1[2:len(sha1)])

	linkFile, err := os.Open(linkPath)
	if err != nil {
		return nil, err
	}

	link, err := DecodeLink(linkFile)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func DecodeLink(reader io.Reader) (*Link, error) {
	link := &Link{}

	m := make(map[string]string)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) == 0 {
			continue
		}

		parts := strings.SplitN(text, " ", 2)
		key := parts[0]
		m[key] = parts[1]
	}

	oid, ok := m["oid"]
	if !ok {
		return nil, errors.New("No Oid in link file")
	}

	link.Oid = oid
	link.Name = m["name"]
	return link, nil
}
