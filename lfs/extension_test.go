package lfs

import (
	"testing"

	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

func TestSortExtensions(t *testing.T) {
	m := map[string]Extension{
		"baz": Extension{
			"baz",
			"baz-clean %f",
			"baz-smudge %f",
			2,
		},
		"foo": Extension{
			"foo",
			"foo-clean %f",
			"foo-smudge %f",
			0,
		},
		"bar": Extension{
			"bar",
			"bar-clean %f",
			"bar-smudge %f",
			1,
		},
	}

	names := []string{"foo", "bar", "baz"}

	sorted, err := SortExtensions(m)

	assert.Equal(t, err, nil)

	for i, ext := range sorted {
		name := names[i]
		assert.Equal(t, name, ext.Name)
		assert.Equal(t, name+"-clean %f", ext.Clean)
		assert.Equal(t, name+"-smudge %f", ext.Smudge)
		assert.Equal(t, i, ext.Priority)
	}
}

func TestSortExtensionsDuplicatePriority(t *testing.T) {
	m := map[string]Extension{
		"foo": Extension{
			"foo",
			"foo-clean %f",
			"foo-smudge %f",
			0,
		},
		"bar": Extension{
			"bar",
			"bar-clean %f",
			"bar-smudge %f",
			0,
		},
	}

	sorted, err := SortExtensions(m)

	assert.NotEqual(t, err, nil)
	assert.Equal(t, len(sorted), 0)
}
