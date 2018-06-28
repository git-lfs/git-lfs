package lfs

import (
	"fmt"
	"strings"

	"github.com/git-lfs/git-lfs/git"
)

// Attribute wraps the structure and some operations of Git's conception of an
// "attribute", as defined here: http://git-scm.com/docs/gitattributes.
type Attribute struct {
	// The Section of an Attribute refers to the location at which all
	// properties are relative to. For example, for a Section with the value
	// "core", Git will produce something like:
	//
	// [core]
	//	autocrlf = true
	//	...
	Section string

	// The Properties of an Attribute refer to all of the keys and values
	// that define that Attribute.
	Properties map[string]string
	// Previous values of these attributes that can be automatically upgraded
	Upgradeables map[string][]string
}

// FilterOptions serves as an argument to Install().
type FilterOptions struct {
	GitConfig  *git.Configuration
	Force      bool
	Local      bool
	System     bool
	SkipSmudge bool
}

func (o *FilterOptions) Install() error {
	if o.SkipSmudge {
		return skipSmudgeFilterAttribute().Install(o)
	}
	return filterAttribute().Install(o)
}

func (o *FilterOptions) Uninstall() error {
	filterAttribute().Uninstall(o)
	return nil
}

func filterAttribute() *Attribute {
	return &Attribute{
		Section: "filter.lfs",
		Properties: map[string]string{
			"clean":    "git-lfs clean -- %f",
			"smudge":   "git-lfs smudge -- %f",
			"process":  "git-lfs filter-process",
			"required": "true",
		},
		Upgradeables: map[string][]string{
			"clean": []string{
				"git-lfs clean %f",
			},
			"smudge": []string{
				"git-lfs smudge %f",
				"git-lfs smudge --skip %f",
				"git-lfs smudge --skip -- %f",
			},
			"process": []string{
				"git-lfs filter",
				"git-lfs filter --skip",
				"git-lfs filter-process --skip",
			},
		},
	}
}

func skipSmudgeFilterAttribute() *Attribute {
	return &Attribute{
		Section: "filter.lfs",
		Properties: map[string]string{
			"clean":    "git-lfs clean -- %f",
			"smudge":   "git-lfs smudge --skip -- %f",
			"process":  "git-lfs filter-process --skip",
			"required": "true",
		},
		Upgradeables: map[string][]string{
			"clean": []string{
				"git-lfs clean -- %f",
			},
			"smudge": []string{
				"git-lfs smudge %f",
				"git-lfs smudge --skip %f",
				"git-lfs smudge -- %f",
			},
			"process": []string{
				"git-lfs filter",
				"git-lfs filter --skip",
				"git-lfs filter-process",
			},
		},
	}
}

// Install instructs Git to set all keys and values relative to the root
// location of this Attribute. For any particular key/value pair, if a matching
// key is already set, it will be overridden if it is either a) empty, or b) the
// `force` argument is passed as true. If an attribute is already set to a
// different value than what is given, and force is false, an error will be
// returned immediately, and the rest of the attributes will not be set.
func (a *Attribute) Install(opt *FilterOptions) error {
	for k, v := range a.Properties {
		var upgradeables []string
		if a.Upgradeables != nil {
			// use pre-normalised key since caller will have set up the same
			upgradeables = a.Upgradeables[k]
		}
		key := a.normalizeKey(k)
		if err := a.set(opt.GitConfig, key, v, upgradeables, opt); err != nil {
			return err
		}
	}

	return nil
}

// normalizeKey makes an absolute path out of a partial relative one. For a
// relative path of "foo", and a root Section of "bar", "bar.foo" will be returned.
func (a *Attribute) normalizeKey(relative string) string {
	return strings.Join([]string{a.Section, relative}, ".")
}

// set attempts to set a single key/value pair portion of this Attribute. If a
// matching key already exists and the value is not equal to the desired value,
// an error will be thrown if force is set to false. If force is true, the value
// will be overridden.
func (a *Attribute) set(gitConfig *git.Configuration, key, value string, upgradeables []string, opt *FilterOptions) error {
	var currentValue string
	if opt.Local {
		currentValue = gitConfig.FindLocal(key)
	} else if opt.System {
		currentValue = gitConfig.FindSystem(key)
	} else {
		currentValue = gitConfig.FindGlobal(key)
	}

	if opt.Force || shouldReset(currentValue, upgradeables) {
		var err error
		if opt.Local {
			_, err = gitConfig.SetLocal(key, value)
		} else if opt.System {
			_, err = gitConfig.SetSystem(key, value)
		} else {
			_, err = gitConfig.SetGlobal(key, value)
		}
		return err
	} else if currentValue != value {
		return fmt.Errorf("The %q attribute should be %q but is %q",
			key, value, currentValue)
	}

	return nil
}

// Uninstall removes all properties in the path of this property.
func (a *Attribute) Uninstall(opt *FilterOptions) {
	if opt.Local {
		opt.GitConfig.UnsetLocalSection(a.Section)
	} else if opt.System {
		opt.GitConfig.UnsetSystemSection(a.Section)
	} else {
		opt.GitConfig.UnsetGlobalSection(a.Section)
	}
}

// shouldReset determines whether or not a value is resettable given its current
// value on the system. If the value is empty (length = 0), then it will pass.
// It will also pass if it matches any upgradeable value
func shouldReset(value string, upgradeables []string) bool {
	if len(value) == 0 {
		return true
	}

	for _, u := range upgradeables {
		if value == u {
			return true
		}
	}

	return false
}
