package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeNone(t *testing.T) {
	evaluate(t, "A:\\some\\regular\\windows\\path", "A:\\some\\regular\\windows\\path")
}

func TestDecodeSingle(t *testing.T) {
	evaluate(t, "A:\\bl\\303\\204\\file.txt", "A:\\blÄ\\file.txt")
}

func TestDecodeMultiple(t *testing.T) {
	evaluate(t, "A:\\fo\\130\\file\\303\\261.txt", "A:\\fo\130\\file\303\261.txt")
}

func evaluate(t *testing.T, input string, expected string) {

	fs := Filesystem{}
	output := fs.DecodePathname(input)

	if output != expected {
		t.Errorf("Expecting same path, got: %s, want: %s.", output, expected)
	}

}

func TestRepositoryPermissions(t *testing.T) {
	m := map[os.FileMode]os.FileMode{
		0777: 0666,
		0755: 0644,
		0700: 0600,
	}
	for k, v := range m {
		fs := Filesystem{repoPerms: v}
		assert.Equal(t, k, fs.RepositoryPermissions(true))
		assert.Equal(t, v, fs.RepositoryPermissions(false))
	}
}

func TestRemoveEmptyObjectDirs(t *testing.T) {
	fs := Filesystem{LFSStorageDir: t.TempDir(), repoPerms: 0644}

	retained := "aabb" + strings.Repeat("0", 60)
	sameDir := "aabb" + strings.Repeat("1", 60)
	sameParent := "aacc" + strings.Repeat("2", 60)
	unique := "ddee" + strings.Repeat("3", 60)

	for _, oid := range []string{retained, sameDir, sameParent, unique} {
		path, err := fs.ObjectPath(oid)
		require.Nil(t, err)
		require.Nil(t, os.WriteFile(path, []byte(oid), 0644))
	}
	for _, oid := range []string{sameDir, sameParent, unique} {
		require.Nil(t, os.Remove(fs.ObjectPathname(oid)))
		fs.RemoveEmptyObjectDirs(oid)
	}

	objectDir := fs.LFSObjectDir()
	assert.FileExists(t, fs.ObjectPathname(retained))
	assert.DirExists(t, filepath.Join(objectDir, "aa", "bb"))
	assert.NoDirExists(t, filepath.Join(objectDir, "aa", "cc"))
	assert.NoDirExists(t, filepath.Join(objectDir, "dd"))
	assert.DirExists(t, objectDir)
}
