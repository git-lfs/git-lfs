package commands

import (
	"github.com/bmizerany/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPointerWithBuildAndCompareStdinMismatch(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	pointer := "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868\n" +
		"size 123"

	cmd := repo.Command("pointer", "--file=some/file", "--stdin")
	cmd.Unsuccessful = true
	cmd.Input = strings.NewReader(pointer + "\n")
	cmd.Before(func() {
		path := filepath.Join(repo.Path, "some")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(filepath.Join(path, "file"), []byte("simple\n"), 0755))
	})

	cmd.Output = "Git LFS pointer for some/file\n\n" +
		"version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868\n" +
		"size 7\n\n" +
		"Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee\n\n" +
		"Pointer from STDIN\n\n" +
		pointer + "\n\n" +
		"Git blob OID: 905bcc24b5dc074ab870f9944178e398eec3b470\n\n" +
		"Pointers do not match"
}

func TestPointerWithBuildAndCompareStdin(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	pointer := "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868\n" +
		"size 7"

	cmd := repo.Command("pointer", "--file=some/file", "--stdin")
	cmd.Input = strings.NewReader(pointer + "\n")
	cmd.Before(func() {
		path := filepath.Join(repo.Path, "some")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(filepath.Join(path, "file"), []byte("simple\n"), 0755))
	})

	cmd.Output = "Git LFS pointer for some/file\n\n" +
		pointer + "\n\n" +
		"Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee\n\n" +
		"Pointer from STDIN\n\n" +
		pointer + "\n\n" +
		"Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee"
}

func TestPointerWithCompareStdin(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	pointer := "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868\n" +
		"size 7"

	cmd := repo.Command("pointer", "--stdin")
	cmd.Input = strings.NewReader(pointer + "\n")
	cmd.Output = "Pointer from STDIN\n\n" + pointer
}

func TestPointerWithInvalidCompareStdin(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("pointer", "--stdin")
	cmd.Output = "Pointer from STDIN\n\nEOF"
	cmd.Unsuccessful = true
}

func TestPointerWithNonPointerCompareStdin(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("pointer", "--stdin")
	cmd.Input = strings.NewReader("not a pointer")
	cmd.Output = "Pointer from STDIN\n\nNot a valid Git LFS pointer file."
	cmd.Unsuccessful = true
}

func TestPointerWithBuildAndCompareMismatch(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	pointer := "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868\n" +
		"size 123"

	cmd := repo.Command("pointer", "--file=some/file", "--pointer=some/pointer")
	cmd.Unsuccessful = true
	cmd.Before(func() {
		path := filepath.Join(repo.Path, "some")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(filepath.Join(path, "pointer"), []byte(pointer+"\n"), 0755))
		assert.Equal(t, nil, ioutil.WriteFile(filepath.Join(path, "file"), []byte("simple\n"), 0755))
	})

	cmd.Output = "Git LFS pointer for some/file\n\n" +
		"version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868\n" +
		"size 7\n\n" +
		"Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee\n\n" +
		"Pointer from some/pointer\n\n" +
		pointer + "\n\n" +
		"Git blob OID: 905bcc24b5dc074ab870f9944178e398eec3b470\n\n" +
		"Pointers do not match"
}

func TestPointerWithBuildAndCompare(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	pointer := "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868\n" +
		"size 7"

	cmd := repo.Command("pointer", "--file=some/file", "--pointer=some/pointer")
	cmd.Before(func() {
		path := filepath.Join(repo.Path, "some")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(filepath.Join(path, "pointer"), []byte(pointer+"\n"), 0755))
		assert.Equal(t, nil, ioutil.WriteFile(filepath.Join(path, "file"), []byte("simple\n"), 0755))
	})

	cmd.Output = "Git LFS pointer for some/file\n\n" +
		pointer + "\n\n" +
		"Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee\n\n" +
		"Pointer from some/pointer\n\n" +
		pointer + "\n\n" +
		"Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee"
}

func TestPointerWithCompare(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	pointer := "version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868\n" +
		"size 7"

	cmd := repo.Command("pointer", "--pointer=some/pointer")
	cmd.Before(func() {
		path := filepath.Join(repo.Path, "some")
		filename := filepath.Join(path, "pointer")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(filename, []byte(pointer+"\n"), 0755))
	})

	cmd.Output = "Pointer from some/pointer\n\n" + pointer
}

func TestPointerWithInvalidCompare(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("pointer", "--pointer=some/pointer")
	cmd.Output = "open some/pointer: no such file or directory"
	cmd.Unsuccessful = true
}

func TestPointerWithNonPointerCompare(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("pointer", "--pointer=some/pointer")
	cmd.Before(func() {
		path := filepath.Join(repo.Path, "some")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(filepath.Join(path, "pointer"), []byte("not a pointer"), 0755))
	})

	cmd.Output = "Pointer from some/pointer\n\nNot a valid Git LFS pointer file."
	cmd.Unsuccessful = true
}

func TestPointerWithBuild(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("pointer", "--file=some/file")
	cmd.Before(func() {
		path := filepath.Join(repo.Path, "some")
		filename := filepath.Join(path, "file")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(filename, []byte("simple\n"), 0755))
	})

	cmd.Output = "Git LFS pointer for some/file\n\n" +
		"version https://git-lfs.github.com/spec/v1\n" +
		"oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868\n" +
		"size 7"
}

func TestPointerWithInvalidBuild(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("pointer", "--file=some/file")
	cmd.Output = "open some/file: no such file or directory"
	cmd.Unsuccessful = true
}

func TestPointerWithoutArgs(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("pointer")
	cmd.Output = "Nothing to do!"
	cmd.Unsuccessful = true
}
