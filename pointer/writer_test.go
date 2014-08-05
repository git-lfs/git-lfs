package pointer

import (
	"github.com/bmizerany/assert"
	"github.com/github/git-media/gitmedia"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteConsistentFile(t *testing.T) {
	path, close := SetupConsistentWriter()
	defer close()

	filename := filepath.Join(path, "valid")
	file, err := newFile(filename, "e9058ab198f6908f702111b0c0fb5b36f99d00554521886c40e2891b349dc7a1")
	if err != nil {
		t.Errorf("file error: %s", err.Error())
	}
	assert.Equal(t, nil, err)

	n, err := file.Write([]byte("yo"))
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, n)

	err = file.Close()
	assert.Equal(t, nil, err)

	by, err := ioutil.ReadFile(filename)
	assert.Equal(t, nil, err)
	assert.Equal(t, "yo", string(by))
}

func TestAttemptWriteToExistingFile(t *testing.T) {
	path, close := SetupConsistentWriter()
	defer close()

	filename := filepath.Join(path, "existing")
	err := ioutil.WriteFile(filename, []byte("yo"), 0777)
	assert.Equal(t, nil, err)

	_, err = newFile(filename, "sha")
	if err == nil {
		t.Fatalf("Expected error!")
	}

	if !strings.Contains(err.Error(), "File exists") {
		t.Fatalf("No problem trying to write to %s", filename)
	}
}

func TestAttemptWriteWithInvalidSHA(t *testing.T) {
	path, close := SetupConsistentWriter()
	defer close()

	filename := filepath.Join(path, "invalid-sha")
	file, err := newFile(filename, "sha")
	assert.Equal(t, nil, err)

	n, err := file.Write([]byte("yo"))
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, n)

	err = file.Close()
	if !strings.Contains(err.Error(), "Unexpected SHA-256") {
		t.Fatalf("No problem trying to write to %s", filename)
	}

	stat, err := os.Stat(filename)
	if err == nil {
		t.Fatalf(".git media file should not exist: %s", filename)
	}
	assert.Equal(t, nil, stat)
}

func SetupConsistentWriter() (string, func()) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(wd, "test")
	oldTempDir := gitmedia.TempDir
	gitmedia.TempDir = filepath.Join(path, "tmp")
	err = os.MkdirAll(path, 0777)
	if err != nil {
		panic(err)
	}

	return path, func() {
		os.RemoveAll(path)
		gitmedia.ResetTempDir()
		gitmedia.TempDir = oldTempDir
	}
}
