package gitattributes

import "github.com/git-lfs/wildmatch"

type Pattern struct {
	wm *wildmatch.Wildmatch
}

func NewPattern(p string) *Pattern {
	return &Pattern{
		wm: wildmatch.NewWildmatch(p, wildmatch.SystemCase),
	}
}

func (p *Pattern) Match(t string) bool {
	return p.wm.Match(t)
}

func (p *Pattern) String() string {
	return p.wm.String()
}
