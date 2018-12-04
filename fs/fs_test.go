package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
