package pointer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/github/git-media/git"
	"github.com/github/git-media/gitmedia"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	linkTemplate = `oid %s
name %s
`
)

// Link provides a link between a git sha1 and a git media oid. Link files
// are stored under .git/media/objects in a structure similar to that of
// .git/objects. The link file contains the git media oid and the name of
// file if it was available when created.
type Link struct {
	Oid  string
	Name string
}

// CreateLink will create a link file from a Pointer. The filename passed
// in is what will be stored in the link contents.
func (p *Pointer) CreateLink(filename string) error {
	hash, err := git.HashObject([]byte(p.Encoded()))
	if err != nil {
		return err
	}

	linkFile, err := gitmedia.LocalLinkPath(hash)
	if err != nil {
		return err
	}
	linkLock := linkFile + ".lock"

	file, err := os.OpenFile(linkLock, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}

	_, err = file.WriteString(fmt.Sprintf(linkTemplate, p.Oid, filename))
	if err != nil {
		os.Remove(linkLock)
		return err
	}

	file.Close()

	os.Remove(linkFile) // Remove the link file if it already existed

	return os.Rename(linkLock, linkFile)
}

// FindLink takes a git sha1 and attempts to find the link file associated
// with it. While this can be used to determine whether the git object is
// a git media object, it should not be considered authoritative. While a link
// file's presence is an indicator that it is a git media file, its absence does
// not mean that it is not a git media file.
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
