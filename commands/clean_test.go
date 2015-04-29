package commands

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestCleanSmallFile(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("clean")
	cmd.Input = bytes.NewBufferString("whatever\n")
	cmd.Output = "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411\n" +
		"size 9"

	path := filepath.Join(repo.Path, ".git", "lfs", "objects")
	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.Before(func() {
		_, err := os.Open(path)
		if _, ok := err.(*os.PathError); !ok {
			t.Fatalf("'%s' should not exist", path)
		}
	})

	file := filepath.Join(path, "cd", "29", "cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411")
	cmd.After(func() {
		by, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		contents := string(by)
		if contents != "whatever\n" {
			t.Fatalf("wrong contents: '%v'", contents)
		}

		by, err = ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatal(err)
		}

		pattern := "git lfs push"
		if !bytes.Contains(by, []byte(pattern)) {
			t.Errorf("hook does not contain %s:\n%s", pattern, string(by))
		}
	})
}

func TestCleanBigFile(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("clean")
	cmd.Input = randomReader(1024)
	cmd.Output = "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5\n" +
		"size 1024"

	path := filepath.Join(repo.Path, ".git", "lfs", "objects")

	cmd.Before(func() {
		_, err := os.Open(path)
		if _, ok := err.(*os.PathError); !ok {
			t.Fatalf("'%s' should not exist", path)
		}
	})

	file := filepath.Join(path, "7c", "d8", "7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5")
	cmd.After(func() {
		stat, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}

		if fileSize := stat.Size(); fileSize != 1024 {
			t.Fatalf("bad size: %d", fileSize)
		}
	})
}

func TestCleanPointerWithExtra(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("clean")
	cmd.Input = bytes.NewBufferString("version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5\n" +
		"size 1024\n\n" +
		"This is my test pointer.  There are many like it, but this one is mine.")
	cmd.Output = "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:f39333aa7ce293ce27742843a1a1cd6e958eaf6b44cfca9039e6b461088df5ba\n" +
		"size 201"

	path := filepath.Join(repo.Path, ".git", "lfs", "objects")

	cmd.Before(func() {
		_, err := os.Open(path)
		if _, ok := err.(*os.PathError); !ok {
			t.Fatalf("'%s' should not exist", path)
		}
	})

	file := filepath.Join(path, "f3", "93", "f39333aa7ce293ce27742843a1a1cd6e958eaf6b44cfca9039e6b461088df5ba")
	cmd.After(func() {
		stat, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}

		if fileSize := stat.Size(); fileSize != 201 {
			t.Fatalf("bad size: %d", fileSize)
		}
	})
}

func TestCleanPointerWithWhitespaceAndExtra(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	head := bytes.NewBufferString("version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5\n" +
		"size 1024\n" +
		// extra white space to fill the initial buffer in DecodeFrom()
		"\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")

	cmd := repo.Command("clean")
	cmd.Input = io.MultiReader(head, randomReader(1024))
	cmd.Output = "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:494f6e07ed15b2e31ca60b31ed8c503bd7553dfd4a76ec73692f766bc67f0bec\n" +
		"size 1537"

	path := filepath.Join(repo.Path, ".git", "lfs", "objects")

	cmd.Before(func() {
		_, err := os.Open(path)
		if _, ok := err.(*os.PathError); !ok {
			t.Fatalf("'%s' should not exist", path)
		}
	})

	file := filepath.Join(path, "49", "4f", "494f6e07ed15b2e31ca60b31ed8c503bd7553dfd4a76ec73692f766bc67f0bec")
	cmd.After(func() {
		stat, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}

		if fileSize := stat.Size(); fileSize != 1537 {
			t.Fatalf("bad size: %d", fileSize)
		}
	})
}

func TestCleanPointer(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	versions := []string{
		"http://git-media.io/v/2",
		"https://hawser.github.com/spec/v1",
		"https://git-lfs.github.com/spec/v1",
	}

	for _, v := range versions {
		cmd := repo.Command("clean")
		cmd.Input = bytes.NewBufferString("version " + v + "\n" +
			"oid sha256:cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411\n" +
			"size 9\n")
		cmd.Output = "version " + v + "\n" +
			"oid sha256:cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411\n" +
			"size 9"

		path := filepath.Join(repo.Path, ".git", "lfs", "objects")

		cmd.Before(func() {
			_, err := os.Open(path)
			if _, ok := err.(*os.PathError); !ok {
				t.Errorf("'%s' should not exist before version %s", path, v)
			}
		})

		cmd.After(func() {
			dirs, err := ioutil.ReadDir(path)
			if _, ok := err.(*os.PathError); ok {
				return // it's ok if this dir does not exist
			}

			if err != nil {
				t.Errorf("Error for version %s: %s", v, err)
				return
			}

			if dirs == nil || len(dirs) == 0 {
				return
			}

			if len(dirs) == 1 && dirs[0].Name() == "logs" {
				return
			}

			for _, dir := range dirs {
				t.Logf("DIR: %v", dir.Name())
			}

			t.Errorf("objects were written to .git/lfs/objects")
		})
	}
}

func TestCleanWithCustomHook(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("clean")
	cmd.Input = bytes.NewBufferString("whatever\n")
	cmd.Output = "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411\n" +
		"size 9"

	path := filepath.Join(repo.Path, ".git", "lfs", "objects")
	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")
	customHook := []byte("test")

	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, customHook, 0755)
		if err != nil {
			t.Fatal(err)
		}

		_, err = os.Open(path)
		if _, ok := err.(*os.PathError); !ok {
			t.Fatalf("'%s' should not exist", path)
		}
	})

	file := filepath.Join(path, "cd", "29", "cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411")
	cmd.After(func() {
		by, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		contents := string(by)
		if contents != "whatever\n" {
			t.Fatalf("wrong contents: '%v'", contents)
		}

		by, err = ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatal(err)
		}

		if string(by) != string(customHook) {
			t.Logf(string(by))
			t.Errorf("Hook does not contain custom hook")
		}
	})
}

func randomReader(n int64) io.Reader {
	return io.LimitReader(&randomDataMaker{rand.NewSource(42)}, n)
}

type randomDataMaker struct {
	src rand.Source
}

func (r *randomDataMaker) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(r.src.Int63() & 0xff)
	}
	return len(p), nil
}
