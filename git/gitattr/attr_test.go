package gitattr

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLines(t *testing.T) {
	lines, err := ParseLines(strings.NewReader("*.dat filter=lfs"))

	assert.NoError(t, err)
	assert.Len(t, lines, 1)

	assert.Equal(t, lines[0].Pattern.String(), "*.dat")
	assert.Equal(t, lines[0].Attrs[0], &Attr{
		K: "filter", V: "lfs",
	})
}

func TestParseLinesManyAttrs(t *testing.T) {
	lines, err := ParseLines(strings.NewReader(
		"*.dat filter=lfs diff=lfs merge=lfs -text"))

	assert.NoError(t, err)

	assert.Len(t, lines, 1)
	assert.Equal(t, lines[0].Pattern.String(), "*.dat")

	assert.Len(t, lines[0].Attrs, 4)
	assert.Equal(t, lines[0].Attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, lines[0].Attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, lines[0].Attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, lines[0].Attrs[3], &Attr{K: "text", V: "false"})
}

func TestParseLinesManyLines(t *testing.T) {
	lines, err := ParseLines(strings.NewReader(strings.Join([]string{
		"*.dat filter=lfs diff=lfs merge=lfs -text",
		"*.jpg filter=lfs diff=lfs merge=lfs -text",
		"# *.pdf filter=lfs diff=lfs merge=lfs -text",
		"*.png filter=lfs diff=lfs merge=lfs -text"}, "\n")))

	assert.NoError(t, err)

	assert.Len(t, lines, 3)
	assert.Equal(t, lines[0].Pattern.String(), "*.dat")
	assert.Equal(t, lines[1].Pattern.String(), "*.jpg")
	assert.Equal(t, lines[2].Pattern.String(), "*.png")

	assert.Len(t, lines[0].Attrs, 4)
	assert.Equal(t, lines[0].Attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, lines[0].Attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, lines[0].Attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, lines[0].Attrs[3], &Attr{K: "text", V: "false"})

	assert.Len(t, lines[1].Attrs, 4)
	assert.Equal(t, lines[1].Attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, lines[1].Attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, lines[1].Attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, lines[1].Attrs[3], &Attr{K: "text", V: "false"})

	assert.Len(t, lines[1].Attrs, 4)
	assert.Equal(t, lines[1].Attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, lines[1].Attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, lines[1].Attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, lines[1].Attrs[3], &Attr{K: "text", V: "false"})
}

func TestParseLinesUnset(t *testing.T) {
	lines, err := ParseLines(strings.NewReader("*.dat -filter"))

	assert.NoError(t, err)
	assert.Len(t, lines, 1)

	assert.Equal(t, lines[0].Pattern.String(), "*.dat")
	assert.Equal(t, lines[0].Attrs[0], &Attr{
		K: "filter", V: "false",
	})
}

func TestParseLinesUnspecified(t *testing.T) {
	lines, err := ParseLines(strings.NewReader("*.dat !filter"))

	assert.NoError(t, err)
	assert.Len(t, lines, 1)

	assert.Equal(t, lines[0].Pattern.String(), "*.dat")
	assert.Equal(t, lines[0].Attrs[0], &Attr{
		K: "filter", Unspecified: true,
	})
}

func TestParseLinesQuotedPattern(t *testing.T) {
	lines, err := ParseLines(strings.NewReader(
		"\"space *.dat\" filter=lfs"))

	assert.NoError(t, err)
	assert.Len(t, lines, 1)

	assert.Equal(t, lines[0].Pattern.String(), "space *.dat")
	assert.Equal(t, lines[0].Attrs[0], &Attr{
		K: "filter", V: "lfs",
	})
}

func TestParseLinesCommented(t *testing.T) {
	lines, err := ParseLines(strings.NewReader(
		"# \"space *.dat\" filter=lfs"))

	assert.NoError(t, err)
	assert.Len(t, lines, 0)
}

func TestParseLinesUnbalancedQuotes(t *testing.T) {
	const text = "\"space *.dat filter=lfs"
	lines, err := ParseLines(strings.NewReader(text))

	assert.Empty(t, lines)
	assert.EqualError(t, err, fmt.Sprintf(
		"git/gitattr: unbalanced quote: %s", text))
}

func TestParseLinesWithNoAttributes(t *testing.T) {
	lines, err := ParseLines(strings.NewReader("*.dat"))

	assert.Len(t, lines, 1)
	assert.NoError(t, err)

	assert.Equal(t, lines[0].Pattern.String(), "*.dat")
	assert.Empty(t, lines[0].Attrs)
}
