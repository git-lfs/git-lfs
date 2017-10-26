package githistory

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/git/githistory/log"
	"github.com/git-lfs/git-lfs/tools"
)

// refUpdater is a type responsible for moving references from one point in the
// Git object graph to another.
type refUpdater struct {
	// CacheFn is a function that returns the SHA1 transformation from an
	// original hash to a new one. It specifies a "bool" return value
	// signaling whether or not that given "old" SHA1 was migrated.
	CacheFn func(old []byte) ([]byte, bool)
	// Logger logs the progress of reference updating.
	Logger *log.Logger
	// Refs is a set of *git.Ref's to migrate.
	Refs []*git.Ref
	// Root is the given directory on disk in which the repository is
	// located.
	Root string
}

// UpdateRefs performs the reference update(s) from existing locations (see:
// Refs) to their respective new locations in the graph (see CacheFn).
//
// It creates reflog entries as well as stderr log entries as it progresses
// through the reference updates.
//
// It returns any error encountered, or nil if the reference update(s) was/were
// successful.
func (r *refUpdater) UpdateRefs() error {
	list := r.Logger.List("migrate: Updating refs")
	defer list.Complete()

	var maxNameLen int
	for _, ref := range r.Refs {
		maxNameLen = tools.MaxInt(maxNameLen, len(ref.Name))
	}

	for _, ref := range r.Refs {
		sha1, err := hex.DecodeString(ref.Sha)
		if err != nil {
			return errors.Wrapf(err, "could not decode: %q", ref.Sha)
		}

		to, ok := r.CacheFn(sha1)
		if !ok {
			continue
		}

		if err := git.UpdateRefIn(r.Root, ref, to, ""); err != nil {
			return err
		}

		namePadding := tools.MaxInt(maxNameLen-len(ref.Name), 0)
		list.Entry(fmt.Sprintf("  %s%s\t%s -> %x", ref.Name, strings.Repeat(" ", namePadding), ref.Sha, to))
	}

	return nil
}
