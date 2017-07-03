package odb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommitReturnsCorrectObjectType(t *testing.T) {
	assert.Equal(t, CommitObjectType, new(Commit).Type())
}

func TestCommitEncoding(t *testing.T) {
	author := &Signature{Name: "John Doe", Email: "john@example.com", When: time.Now()}
	committer := &Signature{Name: "Jane Doe", Email: "jane@example.com", When: time.Now()}

	c := &Commit{
		Author:    author.String(),
		Committer: committer.String(),
		ParentIDs: [][]byte{
			[]byte("aaaaaaaaaaaaaaaaaaaa"), []byte("bbbbbbbbbbbbbbbbbbbb"),
		},
		TreeID: []byte("cccccccccccccccccccc"),
		ExtraHeaders: []*ExtraHeader{
			{"foo", "bar"},
		},
		Message: "initial commit",
	}

	buf := new(bytes.Buffer)

	_, err := c.Encode(buf)
	assert.Nil(t, err)

	assertLine(t, buf, "tree 6363636363636363636363636363636363636363")
	assertLine(t, buf, "parent 6161616161616161616161616161616161616161")
	assertLine(t, buf, "parent 6262626262626262626262626262626262626262")
	assertLine(t, buf, "author %s", author.String())
	assertLine(t, buf, "committer %s", committer.String())
	assertLine(t, buf, "foo bar")
	assertLine(t, buf, "")
	assertLine(t, buf, "initial commit")

	assert.Equal(t, 0, buf.Len())
}

func TestCommitDecoding(t *testing.T) {
	author := &Signature{Name: "John Doe", Email: "john@example.com", When: time.Now()}
	committer := &Signature{Name: "Jane Doe", Email: "jane@example.com", When: time.Now()}

	p1 := []byte("aaaaaaaaaaaaaaaaaaaa")
	p2 := []byte("bbbbbbbbbbbbbbbbbbbb")
	treeId := []byte("cccccccccccccccccccc")

	from := new(bytes.Buffer)
	fmt.Fprintf(from, "author %s\n", author)
	fmt.Fprintf(from, "committer %s\n", committer)
	fmt.Fprintf(from, "parent %s\n", hex.EncodeToString(p1))
	fmt.Fprintf(from, "parent %s\n", hex.EncodeToString(p2))
	fmt.Fprintf(from, "foo bar\n")
	fmt.Fprintf(from, "tree %s\n", hex.EncodeToString(treeId))
	fmt.Fprintf(from, "\ninitial commit\n")

	flen := from.Len()

	commit := new(Commit)
	n, err := commit.Decode(from, int64(flen))

	assert.Nil(t, err)
	assert.Equal(t, flen, n)

	assert.Equal(t, author.String(), commit.Author)
	assert.Equal(t, committer.String(), commit.Committer)
	assert.Equal(t, [][]byte{p1, p2}, commit.ParentIDs)
	assert.Equal(t, 1, len(commit.ExtraHeaders))
	assert.Equal(t, "foo", commit.ExtraHeaders[0].K)
	assert.Equal(t, "bar", commit.ExtraHeaders[0].V)
	assert.Equal(t, "initial commit", commit.Message)
}

func TestCommitDecodingWithMessageKeywordPrefix(t *testing.T) {
	author := &Signature{Name: "John Doe", Email: "john@example.com", When: time.Now()}
	committer := &Signature{Name: "Jane Doe", Email: "jane@example.com", When: time.Now()}

	treeId := []byte("aaaaaaaaaaaaaaaaaaaa")
	treeIdAscii := hex.EncodeToString(treeId)

	from := new(bytes.Buffer)
	fmt.Fprintf(from, "author %s\n", author)
	fmt.Fprintf(from, "committer %s\n", committer)
	fmt.Fprintf(from, "tree %s\n", hex.EncodeToString(treeId))
	fmt.Fprintf(from, "\ntree <- initial commit\n")

	flen := from.Len()

	commit := new(Commit)
	n, err := commit.Decode(from, int64(flen))

	assert.NoError(t, err)
	assert.Equal(t, flen, n)

	assert.Equal(t, author.String(), commit.Author)
	assert.Equal(t, committer.String(), commit.Committer)
	assert.Equal(t, treeIdAscii, hex.EncodeToString(commit.TreeID))
	assert.Equal(t, "tree <- initial commit", commit.Message)
}

func assertLine(t *testing.T, buf *bytes.Buffer, wanted string, args ...interface{}) {
	got, err := buf.ReadString('\n')

	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf(wanted, args...), strings.TrimSuffix(got, "\n"))
}
