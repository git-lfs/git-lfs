package githistory

import (
	"testing"

	"github.com/git-lfs/git-lfs/git"
	"github.com/stretchr/testify/assert"
)

func TestRefUpdaterMovesRefs(t *testing.T) {
	db := DatabaseFromFixture(t, "linear-history-with-tags.git")
	root, _ := db.Root()

	updater := &refUpdater{
		CacheFn: func(old []byte) ([]byte, bool) {
			return HexDecode(t, "d941e4756add6b06f5bee766fcf669f55419f13f"), true
		},
		Refs: []*git.Ref{
			{
				Name: "middle",
				Sha:  "228afe30855933151f7a88e70d9d88314fd2f191",
				Type: git.RefTypeLocalTag,
			},
		},
		Root: root,
	}

	err := updater.UpdateRefs()

	assert.NoError(t, err)

	AssertRef(t, db,
		"refs/tags/middle", HexDecode(t, "d941e4756add6b06f5bee766fcf669f55419f13f"))
}

func TestRefUpdaterIgnoresUnovedRefs(t *testing.T) {
	db := DatabaseFromFixture(t, "linear-history-with-tags.git")
	root, _ := db.Root()

	updater := &refUpdater{
		CacheFn: func(old []byte) ([]byte, bool) {
			return nil, false
		},
		Refs: []*git.Ref{
			{
				Name: "middle",
				Sha:  "228afe30855933151f7a88e70d9d88314fd2f191",
				Type: git.RefTypeLocalTag,
			},
		},
		Root: root,
	}

	err := updater.UpdateRefs()

	assert.NoError(t, err)

	AssertRef(t, db,
		"refs/tags/middle", HexDecode(t, "228afe30855933151f7a88e70d9d88314fd2f191"))
}
