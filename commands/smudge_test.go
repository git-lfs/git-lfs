package commands

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmizerany/assert"
)

func TestSmudge(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")
	progressFile := filepath.Join(repo.Path, ".git", "progress")

	// simple smudge example
	cmd := repo.Command("smudge", "somefile")
	cmd.Input = bytes.NewBufferString("version https://git-lfs.github.com/spec/v1\noid sha256:SOMEOID\nsize 7\n")
	cmd.Output = "simple"
	cmd.Env = append(cmd.Env, "GIT_LFS_PROGRESS="+progressFile)

	cmd.Before(func() {
		path := filepath.Join(repo.Path, ".git", "lfs", "objects", "SO", "ME")
		file := filepath.Join(path, "SOMEOID")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(file, []byte("simple\n"), 0755))
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
		assert.Equal(t, 2, len(progLines))
		assert.Equal(t, "smudge 1/1 7/7 somefile", string(progLines[0]))
		assert.Equal(t, "", string(progLines[1]))
	})

	// smudge with custom hook
	cmd = repo.Command("smudge")
	cmd.Input = bytes.NewBufferString("version https://git-lfs.github.com/spec/v1\noid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393\nsize 9")
	cmd.Output = "whatever"
	customHook := "echo 'yo'"

	cmd.Before(func() {
		path := filepath.Join(repo.Path, ".git", "lfs", "objects", "4d", "7a")
		file := filepath.Join(path, "4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393")
		repo.WriteFile(file, "whatever\n")
		repo.WriteFile(prePushHookFile, customHook)
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

	mediaPath := filepath.Join(repo.Path, ".git", "lfs", "objects", "4d", "7a")
	mediaFile := filepath.Join(mediaPath, "4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393")

	// smudge --info with Git LFS pointer format, without local file
	cmd := repo.Command("smudge", "--info")
	cmd.Input = bytes.NewBufferString("version https://git-lfs.github.com/spec/v1\noid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393\nsize 123\n")
	cmd.Output = "123 --"

	// smudge --info with Git LFS pointer format, with local file
	cmd = repo.Command("smudge", "--info")
	cmd.Input = bytes.NewBufferString("version https://git-lfs.github.com/spec/v1\noid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393\nsize 123\n")
	cmd.Output = "9 " + mediaFile

	cmd.Before(func() {
		assert.Equal(t, nil, os.MkdirAll(mediaPath, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(mediaFile, []byte("whatever\n"), 0755))
	})
}

func TestSmudgePassesInvalidData(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("smudge", "somefile")
	cmd.Input = bytes.NewBufferString("this is not a pointer file\n")
	cmd.Output = "this is not a pointer file"
}
