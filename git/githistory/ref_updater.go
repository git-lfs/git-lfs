package githistory

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/tasklog"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/git-lfs/gitobj/v2"
)

// refUpdater is a type responsible for moving references from one point in the
// Git object graph to another.
type refUpdater struct {
	// cacheFn is a function that returns the SHA1 transformation from an
	// original hash to a new one. It specifies a "bool" return value
	// signaling whether or not that given "old" SHA1 was migrated.
	cacheFn func(old []byte) ([]byte, bool)
	// logger logs the progress of reference updating.
	logger *tasklog.Logger
	// refs is a set of *git.Ref's to migrate.
	refs []*git.Ref
	// root is the given directory on disk in which the repository is
	// located.
	root string
	// db is the *ObjectDatabase from which blobs, commits, and trees are
	// loaded.
	db *gitobj.ObjectDatabase
}

// updateRefs performs the reference update(s) from existing locations (see:
// Refs) to their respective new locations in the graph (see CacheFn).
//
// It creates reflog entries as well as stderr log entries as it progresses
// through the reference updates.
//
// It returns any error encountered, or nil if the reference update(s) was/were
// successful.
func (r *refUpdater) updateRefs() error {
	list := r.logger.List(tr.Tr.Get("Updating refs"))
	defer list.Complete()

	var maxNameLen int
	for _, ref := range r.refs {
		maxNameLen = max(maxNameLen, len(ref.Name))
	}

	seen := make(map[string]struct{})
	for _, ref := range r.refs {
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
		return nil, errors.Wrap(err, tr.Tr.Get("could not rewrite tag: %s", tag.Name))
	}
	return newTag, nil
}

func (r *refUpdater) updateOneRef(list *tasklog.ListTask, maxNameLen int, seen map[string]struct{}, ref *git.Ref) error {
	sha1, err := hex.DecodeString(ref.Sha)
	if err != nil {
		return errors.Wrap(err, tr.Tr.Get("could not decode: %q", ref.Sha))
	}

	refspec := ref.Refspec()
	if _, ok := seen[refspec]; ok {
		return nil
	}
	seen[refspec] = struct{}{}

	to, ok := r.cacheFn(sha1)

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
				return errors.Wrap(err, tr.Tr.Get("could not decode: %q", ref.Sha))
			}

			newTag, err := r.updateOneTag(tag, updatedSha)
			if newTag == nil {
				return err
			}
			to = newTag
			ok = true
		} else if tag != nil && tag.ObjectType == gitobj.CommitObjectType {
			toObj, okObj := r.cacheFn(tag.Object)
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

	if err := git.UpdateRefIn(r.root, ref, to, ""); err != nil {
		return err
	}

	namePadding := max(maxNameLen-len(ref.Name), 0)
	list.Entry(fmt.Sprintf("  %s%s\t%s -> %x", ref.Name, strings.Repeat(" ", namePadding), ref.Sha, to))
	return nil
}
