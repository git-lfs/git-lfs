package gitattributes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
)

type Tree struct {
	Repo string
}

func (t *Tree) Attributes(name string) (map[string]string, error) {
	applied := make(map[string]string)

	path := unjoin(filepath.Dir(name), t.Repo)
	trees := strings.Split(path, "/")

	for i := 0; i <= len(trees); i++ {
		dir := trees[:i]
		path := join(strings.Join(dir, "/"), ".gitattributes")

		f, err := os.Open(join(t.Repo, path))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		fmt.Println(f.Name())

		entries, err := ParseEntries(f.Name(), f)
		if err != nil {
			return nil, badread(err, f.Name())
		}

		matching := entries.Matching(name)
		for _, match := range matching {
			for _, attr := range match.Attributes {
				if attr.Negated {
					delete(applied, attr.Type)
				} else {
					applied[attr.Type] = attr.Value
				}
			}
		}

		f.Close()
	}

	return applied, nil
}

func badread(err error, name string) error {
	return errors.Wrapf(err, "git/gitattributes: coulnd not read %s", name)
}
