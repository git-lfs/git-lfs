package git

import (
	"bufio"
	"bytes"
	"errors"
	"regexp"
	"strings"
)

var z40 = regexp.MustCompile(`\^?0{40}`)

type GitObject struct {
	Sha1 string
	Name string
}

func RevListObjects(left, right string) ([]*GitObject, error) {
	objects := make([]*GitObject, 0)

	refArgs := []string{"rev-list", "--objects"}

	if left == "" {
		return nil, errors.New("Left commit required")
	}

	refArgs = append(refArgs, left)

	if right != "" && !z40.MatchString(right) {
		refArgs = append(refArgs, right)
	}

	output, err := simpleExec(nil, "git", refArgs...)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewBufferString(output))
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		objects = append(objects, &GitObject{line[0], strings.Join(line[1:len(line)], " ")})
	}

	return objects, nil
}
