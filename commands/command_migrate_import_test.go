package commands

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRegexEvenCount(t *testing.T) {
	regexes := compileRegexList("!", "ASCII|Unicode!postscript!special")

	assert.Equal(t, false, matchAntimatch("binary", regexes))
	assert.Equal(t, true, matchAntimatch("ASCII", regexes))
	assert.Equal(t, false, matchAntimatch("ASCII postscript", regexes))
	assert.Equal(t, true, matchAntimatch("ASCII postscript special", regexes))
}

func TestParseRegexOddCount(t *testing.T) {
	regexes := compileRegexList("!", "ASCII|Unicode!postscript")

	assert.Equal(t, false, matchAntimatch("binary", regexes))
	assert.Equal(t, true, matchAntimatch("ASCII", regexes))
	assert.Equal(t, false, matchAntimatch("ASCII postscript", regexes))
}

var emptyRegexList = compileRegexList(";", "")

func TestSkipBecauseByMatchingTextPath(t *testing.T) {
	reader := strings.NewReader("data")
	outReader, skip := skipBecauseText(reader, "path.txt", compileRegexList(";", "txt$"), emptyRegexList, emptyRegexList)
	assert.Equal(t, reader, outReader, "reader should be unaffected")
	assert.Equal(t, true, skip, "Text path match should skip")
}

func TestSkipBecauseByMatchingBinaryPath(t *testing.T) {
	reader := strings.NewReader("data")
	outReader, skip := skipBecauseText(reader, "path.bin", compileRegexList(";", "txt$"), compileRegexList(";", "bin$"), emptyRegexList)
	assert.Equal(t, reader, outReader, "reader should be unaffected")
	assert.Equal(t, false, skip, "Binary path match should count")
}

func TestSkipBecauseByMatchingTextData(t *testing.T) {
	readerData := "Plain text"
	reader := strings.NewReader(readerData)
	outReader, skip := skipBecauseText(reader, "path/file", emptyRegexList, emptyRegexList, compileRegexList(";", "ASCII"))
	assert.NotEqual(t, reader, outReader, "reader should be different")
	data, err := io.ReadAll(outReader)
	assert.Equal(t, err, nil)
	assert.Equal(t, readerData, bytes.NewBuffer(data).String())
	assert.Equal(t, true, skip, "Text path match should skip")
}

func TestSkipBecauseByMatchingUnicodeData(t *testing.T) {
	readerData := "Unicode £ text"
	reader := strings.NewReader(readerData)
	outReader, skip := skipBecauseText(reader, "path/file", emptyRegexList, emptyRegexList, compileRegexList(";", "Unicode"))
	assert.NotEqual(t, reader, outReader, "reader should be different")
	data, err := io.ReadAll(outReader)
	assert.Equal(t, err, nil)
	assert.Equal(t, readerData, bytes.NewBuffer(data).String())
	assert.Equal(t, true, skip, "Text path match should skip")
}

func TestSkipBecauseByMatchingBinaryData(t *testing.T) {
	readerData := "\xAE\bi\n\a\ry"
	reader := strings.NewReader(readerData)
	outReader, skip := skipBecauseText(reader, "path/file", emptyRegexList, emptyRegexList, compileRegexList(";", "ASCII|Unicode"))
	assert.NotEqual(t, reader, outReader, "reader should be different")
	data, err := io.ReadAll(outReader)
	assert.Equal(t, err, nil)
	assert.Equal(t, readerData, bytes.NewBuffer(data).String())
	assert.Equal(t, false, skip, "Binary path match should count")
}
