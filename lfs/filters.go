package lfs

import (
	"fmt"
	"regexp"

	"github.com/github/git-lfs/git"
)

var (
	valueRegexp = regexp.MustCompile("\\Agit[\\-\\s]media")
)

// A Filter represents a git-configurable attribute of the type "filter". It
// holds a name and a value.
type Filter struct {
	Name  string
	Value string
}

// Install installs the filter in context. It returns any errors it encounters
// along the way. If the value for this filter is set, but not equal to the
// value that is given, then an exception will be returned.  If `force` is
// passed as true in that same situtation, the value will be overridden.  All
// other cases will pass.
func (f *Filter) Install(force bool) error {
	key := fmt.Sprintf("filter.lfs.%s", f.Name)

	currentValue := git.Config.Find(key)
	if force || f.shouldReset(currentValue) {
		git.Config.UnsetGlobal(key)
		git.Config.SetGlobal(key, f.Value)

		return nil
	} else if currentValue != f.Value {
		return fmt.Errorf("The %s filter should be \"%s\" but is \"%s\"",
			f.Name, f.Value, currentValue)
	}

	return nil
}

// shouldReset determines whether or not a value is resettable given its current
// value on the system. If the value is empty (length = 0), then it will pass.
// Otherwise, it will pass if the below regex matches.
func (f *Filter) shouldReset(value string) bool {
	if len(value) == 0 {
		return true
	}

	return valueRegexp.MatchString(value)
}

// The Filters type represents a set of filters to install on the system. It
// exists purely for the convenience of being able to stick the Teardown method
// somewhere where it makes sense.
type Filters []*Filter

// Setup installs all filters in range of the current Filters instance. It
// passes the force argument directly to each of them. If any filter returns an
// error, Setup will halt and return that error.
func (fs *Filters) Setup(force bool) error {
	for _, f := range *fs {
		if err := f.Install(force); err != nil {
			return err
		}
	}

	return nil

}

// Teardown unsets all filters that were installed.
func (fs *Filters) Teardown() {
	git.Config.UnsetGlobalSection("filters.lfs")
}
