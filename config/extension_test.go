package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Nil(t, err)

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

	assert.NotNil(t, err)
	assert.Empty(t, sorted)
}
