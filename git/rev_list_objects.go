package git

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

var z40 = regexp.MustCompile(`\^?0{40}`)

func RevListObjects(left, right string) ([]string, error) {
	objects := make([]string, 0)

	refArgs := []string{"rev-list", "--objects"}

	if left == "" {
		return nil, errors.New("Left commit required")
	}

	refArgs = append(refArgs, left)

	if right != "" && !z40.MatchString(right) {
		refArgs = append(refArgs, right)
	}

	output, err := exec.Command("git", refArgs...).Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(output))
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		objects = append(objects, line[0])
	}

	return objects, nil
}
