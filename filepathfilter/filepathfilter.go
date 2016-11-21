package filepathfilter

import "github.com/git-lfs/git-lfs/tools"

type Filter struct {
	include []string
	exclude []string
}

func New(include, exclude []string) *Filter {
	return &Filter{include: include, exclude: exclude}
}

func (f *Filter) Allows(path string) bool {
	if f == nil {
		return true
	}

	return tools.FilenamePassesIncludeExcludeFilter(path, f.include, f.exclude)
}
