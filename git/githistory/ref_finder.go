package githistory

import (
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/git/odb"
)

type rfopt func(*RefFinder)

var (
	localTypes = []git.RefType{
		git.RefTypeHEAD,
		git.RefTypeLocalBranch,
		git.RefTypeLocalTag,
		git.RefTypeOther,
	}

	RefFinderLocalOnly = func(r *RefFinder) {
		r.AllowedTypes = localTypes
	}
)

type RefFinder struct {
	AllowedTypes []git.RefType

	db *odb.ObjectDatabase
}

func NewRefFinder(db *odb.ObjectDatabase, opts ...rfopt) *RefFinder {
	r := &RefFinder{
		db: db,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *RefFinder) FindRefs() ([]*git.Ref, error) {
	var refs []*git.Ref
	var err error

	if root, ok := r.db.Root(); ok {
		refs, err = git.AllRefsIn(root)
	} else {
		refs, err = git.AllRefs()
	}

	if err != nil {
		return nil, err
	}

	return r.filter(refs), nil
}

func (r *RefFinder) filter(all []*git.Ref) []*git.Ref {
	filtered := make([]*git.Ref, 0, len(all))
	for _, ref := range all {
		if r.allowsType(ref.Type) {
			filtered = append(filtered, ref)
		}
	}

	return filtered
}

func (r *RefFinder) allowsType(t git.RefType) bool {
	for _, allowed := range r.AllowedTypes {
		if allowed == t {
			return true
		}
	}
	return len(r.AllowedTypes) == 0
}
