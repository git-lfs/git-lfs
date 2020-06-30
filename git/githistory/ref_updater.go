package githistory

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tasklog"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/gitobj/v2"
)

// refUpdater is a type responsible for moving references from one point in the
// Git object graph to another.
type refUpdater struct {
	// CacheFn is a function that returns the SHA1 transformation from an
	// original hash to a new one. It specifies a "bool" return value
	// signaling whether or not that given "old" SHA1 was migrated.
	CacheFn func(old []byte) ([]byte, bool)
	// Logger logs the progress of reference updating.
	Logger *tasklog.Logger
	// Refs is a set of *git.Ref's to migrate.
	Refs []*git.Ref
	// Root is the given directory on disk in which the repository is
	// located.
	Root string

	db *gitobj.ObjectDatabase
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

	seen := make(map[string]struct{})
	for _, ref := range r.Refs {
		if err := r.updateOneRef(list, maxNameLen, seen, ref); err != nil {
			return err
		}
	}

	return nil
}

func (r *refUpdater) updateOneTag(tag *gitobj.Tag, toObj []byte) ([]byte, error) {
	newTag, err := r.db.WriteTag(&gitobj.Tag{
		Object:     toObj,
		ObjectType: tag.ObjectType,
		Name:       tag.Name,
		Tagger:     tag.Tagger,

		Message: tag.Message,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "could not rewrite tag: %s", tag.Name)
	}
	return newTag, nil
}

func (r *refUpdater) updateOneRef(list *tasklog.ListTask, maxNameLen int, seen map[string]struct{}, ref *git.Ref) error {
	sha1, err := hex.DecodeString(ref.Sha)
	if err != nil {
		return errors.Wrapf(err, "could not decode: %q", ref.Sha)
	}

	refspec := ref.Refspec()
	if _, ok := seen[refspec]; ok {
		return nil
	}
	seen[refspec] = struct{}{}

	to, ok := r.CacheFn(sha1)

	if ref.Type == git.RefTypeLocalTag {
		tag, _ := r.db.Tag(sha1)
		if tag != nil && tag.ObjectType == gitobj.TagObjectType {
			innerTag, _ := r.db.Tag(tag.Object)
			name := fmt.Sprintf("refs/tags/%s", innerTag.Name)
			if _, ok := seen[name]; !ok {
				old, err := git.ResolveRef(name)
				if err != nil {
					return err
				}

				err = r.updateOneRef(list, maxNameLen, seen, old)
				if err != nil {
					return err
				}
			}

			updated, err := git.ResolveRef(name)
			if err != nil {
				return err
			}
			updatedSha, err := hex.DecodeString(updated.Sha)
			if err != nil {
				return errors.Wrapf(err, "could not decode: %q", ref.Sha)
			}

			newTag, err := r.updateOneTag(tag, updatedSha)
			if newTag == nil {
				return err
			}
			to = newTag
			ok = true
		} else if tag != nil && tag.ObjectType == gitobj.CommitObjectType {
			toObj, okObj := r.CacheFn(tag.Object)
			if !okObj {
				return nil
			}

			newTag, err := r.updateOneTag(tag, toObj)
			if newTag == nil {
				return err
			}
			to = newTag
			ok = true
		}
	}

	if !ok {
		return nil
	}

	if err := git.UpdateRefIn(r.Root, ref, to, ""); err != nil {
		return err
	}

	namePadding := tools.MaxInt(maxNameLen-len(ref.Name), 0)
	list.Entry(fmt.Sprintf("  %s%s\t%s -> %x", ref.Name, strings.Repeat(" ", namePadding), ref.Sha, to))
	return nil
}
