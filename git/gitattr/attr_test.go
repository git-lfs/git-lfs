package gitattr

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLines(t *testing.T) {
	lines, _, err := ParseLines(strings.NewReader("*.dat filter=lfs"))

	assert.NoError(t, err)
	assert.Len(t, lines, 1)

	assert.Implements(t, (*PatternLine)(nil), lines[0])

	assert.Equal(t, lines[0].(PatternLine).Pattern().String(), "*.dat")
	assert.Equal(t, lines[0].Attrs()[0], &Attr{
		K: "filter", V: "lfs",
	})
}

func TestParseLinesManyAttrs(t *testing.T) {
	lines, _, err := ParseLines(strings.NewReader(
		"*.dat filter=lfs diff=lfs merge=lfs -text crlf"))

	assert.NoError(t, err)

	assert.Len(t, lines, 1)

	assert.Implements(t, (*PatternLine)(nil), lines[0])

	assert.Equal(t, lines[0].(PatternLine).Pattern().String(), "*.dat")

	assert.Len(t, lines[0].Attrs(), 5)
	assert.Equal(t, lines[0].Attrs()[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, lines[0].Attrs()[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, lines[0].Attrs()[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, lines[0].Attrs()[3], &Attr{K: "text", V: "false"})
	assert.Equal(t, lines[0].Attrs()[4], &Attr{K: "crlf", V: "true"})
}

func TestParseLinesManyLines(t *testing.T) {
	lines, _, err := ParseLines(strings.NewReader(strings.Join([]string{
		"*.dat filter=lfs diff=lfs merge=lfs -text",
		"*.jpg filter=lfs diff=lfs merge=lfs -text",
		"# *.pdf filter=lfs diff=lfs merge=lfs -text",
		"*.png filter=lfs diff=lfs merge=lfs -text",
		"*.txt text"}, "\n")))

	assert.NoError(t, err)

	assert.Len(t, lines, 4)

	assert.Implements(t, (*PatternLine)(nil), lines[0])
	assert.Implements(t, (*PatternLine)(nil), lines[1])
	assert.Implements(t, (*PatternLine)(nil), lines[2])
	assert.Implements(t, (*PatternLine)(nil), lines[3])

	assert.Equal(t, lines[0].(PatternLine).Pattern().String(), "*.dat")
	assert.Equal(t, lines[1].(PatternLine).Pattern().String(), "*.jpg")
	assert.Equal(t, lines[2].(PatternLine).Pattern().String(), "*.png")
	assert.Equal(t, lines[3].(PatternLine).Pattern().String(), "*.txt")

	assert.Len(t, lines[0].Attrs(), 4)
	assert.Equal(t, lines[0].Attrs()[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, lines[0].Attrs()[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, lines[0].Attrs()[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, lines[0].Attrs()[3], &Attr{K: "text", V: "false"})

	assert.Len(t, lines[1].Attrs(), 4)
	assert.Equal(t, lines[1].Attrs()[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, lines[1].Attrs()[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, lines[1].Attrs()[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, lines[1].Attrs()[3], &Attr{K: "text", V: "false"})

	assert.Len(t, lines[2].Attrs(), 4)
	assert.Equal(t, lines[2].Attrs()[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, lines[2].Attrs()[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, lines[2].Attrs()[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, lines[2].Attrs()[3], &Attr{K: "text", V: "false"})

	assert.Len(t, lines[3].Attrs(), 1)
	assert.Equal(t, lines[3].Attrs()[0], &Attr{K: "text", V: "true"})
}

func TestParseLinesUnset(t *testing.T) {
	lines, _, err := ParseLines(strings.NewReader("*.dat -filter"))

	assert.NoError(t, err)
	assert.Len(t, lines, 1)

	assert.Implements(t, (*PatternLine)(nil), lines[0])

	assert.Equal(t, lines[0].(PatternLine).Pattern().String(), "*.dat")
	assert.Equal(t, lines[0].Attrs()[0], &Attr{
		K: "filter", V: "false",
	})
}

func TestParseLinesUnspecified(t *testing.T) {
	lines, _, err := ParseLines(strings.NewReader("*.dat !filter"))

	assert.NoError(t, err)
	assert.Len(t, lines, 1)

	assert.Implements(t, (*PatternLine)(nil), lines[0])

	assert.Equal(t, lines[0].(PatternLine).Pattern().String(), "*.dat")
	assert.Equal(t, lines[0].Attrs()[0], &Attr{
		K: "filter", Unspecified: true,
	})
}

func TestParseLinesQuotedPattern(t *testing.T) {
	lines, _, err := ParseLines(strings.NewReader(
		"\"space *.dat\" filter=lfs"))

	assert.NoError(t, err)
	assert.Len(t, lines, 1)

	assert.Implements(t, (*PatternLine)(nil), lines[0])

	assert.Equal(t, lines[0].(PatternLine).Pattern().String(), "space *.dat")
	assert.Equal(t, lines[0].Attrs()[0], &Attr{
		K: "filter", V: "lfs",
	})
}

func TestParseLinesCommented(t *testing.T) {
	lines, _, err := ParseLines(strings.NewReader(
		"# \"space *.dat\" filter=lfs"))

	assert.NoError(t, err)
	assert.Len(t, lines, 0)
}

func TestParseLinesUnbalancedQuotes(t *testing.T) {
	const text = "\"space *.dat filter=lfs"
	lines, _, err := ParseLines(strings.NewReader(text))

	assert.Empty(t, lines)
	assert.EqualError(t, err, fmt.Sprintf(
		"unbalanced quote: %s", text))
}

func TestParseLinesWithNoAttributes(t *testing.T) {
	lines, _, err := ParseLines(strings.NewReader("*.dat"))

	assert.Len(t, lines, 1)
	assert.NoError(t, err)

	assert.Implements(t, (*PatternLine)(nil), lines[0])

	assert.Equal(t, lines[0].(PatternLine).Pattern().String(), "*.dat")
	assert.Empty(t, lines[0].Attrs())
}

func TestParseLinesWithMacros(t *testing.T) {
	lines, _, err := ParseLines(strings.NewReader(strings.Join([]string{
		"[attr]lfs filter=lfs diff=lfs merge=lfs -text",
		"*.dat lfs",
		"*.txt text"}, "\n")))

	assert.Len(t, lines, 3)
	assert.NoError(t, err)

	assert.Implements(t, (*MacroLine)(nil), lines[0])
	assert.Implements(t, (*PatternLine)(nil), lines[1])
	assert.Implements(t, (*PatternLine)(nil), lines[2])

	assert.Equal(t, lines[0].(MacroLine).Macro(), "lfs")
	assert.Len(t, lines[0].Attrs(), 4)
	assert.Equal(t, lines[0].Attrs()[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, lines[0].Attrs()[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, lines[0].Attrs()[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, lines[0].Attrs()[3], &Attr{K: "text", V: "false"})

	assert.Equal(t, lines[1].(PatternLine).Pattern().String(), "*.dat")
	assert.Len(t, lines[1].Attrs(), 1)
	assert.Equal(t, lines[1].Attrs()[0], &Attr{K: "lfs", V: "true"})

	assert.Equal(t, lines[2].(PatternLine).Pattern().String(), "*.txt")
	assert.Len(t, lines[2].Attrs(), 1)
	assert.Equal(t, lines[2].Attrs()[0], &Attr{K: "text", V: "true"})
}
