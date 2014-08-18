package commands

import (
	"bytes"
	"github.com/bmizerany/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestSmudge(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")
	progressFile := filepath.Join(repo.Path, ".git", "progress")

	// simple smudge example
	cmd := repo.Command("smudge", "somefile")
	cmd.Input = bytes.NewBufferString("version http://git-media.io/v/2\noid sha256:SOMEOID\nsize 9\n")
	cmd.Output = "whatever"
	cmd.Env = append(cmd.Env, "GIT_MEDIA_PROGRESS="+progressFile)

	cmd.Before(func() {
		path := filepath.Join(repo.Path, ".git", "media", "SO", "ME")
		file := filepath.Join(path, "SOMEOID")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(file, []byte("whatever\n"), 0755))
	})

	cmd.After(func() {
		// assert hook is created
		stat, err := os.Stat(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, stat.IsDir())

		// assert progress file
		progress, err := ioutil.ReadFile(progressFile)
		assert.Equal(t, nil, err)
		progLines := bytes.Split(progress, []byte("\n"))
		assert.Equal(t, 3, len(progLines))
		assert.Equal(t, "smudge 1/1 0 somefile", string(progLines[0]))
		assert.Equal(t, "smudge 1/1 100 somefile", string(progLines[1]))
		assert.Equal(t, "", string(progLines[2]))
	})

	// smudge with custom hook
	cmd = repo.Command("smudge")
	cmd.Input = bytes.NewBufferString("# git-media\nSOMEOID")
	cmd.Output = "whatever"
	customHook := []byte("echo 'yo'")

	cmd.Before(func() {
		path := filepath.Join(repo.Path, ".git", "media", "SO", "ME")
		file := filepath.Join(path, "SOMEOID")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(file, []byte("whatever\n"), 0755))
		assert.Equal(t, nil, ioutil.WriteFile(prePushHookFile, customHook, 0755))
	})

	cmd.After(func() {
		// assert custom hook is not overwritten
		by, err := ioutil.ReadFile(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, string(customHook), string(by))
	})
}

func TestSmudgeInfo(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	mediaPath := filepath.Join(repo.Path, ".git", "media", "SO", "ME")
	mediaFile := filepath.Join(mediaPath, "SOMEOID")

	// smudge --info with old pointer format, without local file
	cmd := repo.Command("smudge", "--info")
	cmd.Input = bytes.NewBufferString("# git-media\nSOMEOID")
	cmd.Output = "0 --"

	// smudge --info with old pointer format, with local file
	cmd = repo.Command("smudge", "--info")
	cmd.Input = bytes.NewBufferString("# git-media\nSOMEOID")
	cmd.Output = "9 " + mediaFile

	cmd.Before(func() {
		assert.Equal(t, nil, os.MkdirAll(mediaPath, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(mediaFile, []byte("whatever\n"), 0755))
	})

	// smudge --info with ini pointer format, without local file
	cmd = repo.Command("smudge", "--info")
	cmd.Input = bytes.NewBufferString("version http://git-media.io/v/2\noid sha256:SOMEOID\nsize 123\n")
	cmd.Output = "123 --"

	// smudge --info with ini pointer format, with local file
	cmd = repo.Command("smudge", "--info")
	cmd.Input = bytes.NewBufferString("version http://git-media.io/v/2\noid sha256:SOMEOID\nsize 123\n")
	cmd.Output = "9 " + mediaFile

	cmd.Before(func() {
		assert.Equal(t, nil, os.MkdirAll(mediaPath, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(mediaFile, []byte("whatever\n"), 0755))
	})
}
