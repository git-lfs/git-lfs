package lfs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/github/git-lfs/git"
)

var (
	valueRegexp = regexp.MustCompile("\\Agit[\\-\\s]media")
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
}

// Install instructs Git to set all keys and values relative to the root
// location of this Attribute. For any particular key/value pair, if a matching
// key is already set, it will be overridden if it is either a) empty, or b) the
// `force` argument is passed as true. If an attribute is already set to a
// different value than what is given, and force is false, an error will be
// returned immediately, and the rest of the attributes will not be set.
func (a *Attribute) Install(force bool) error {
	for k, v := range a.Properties {
		key := a.normalizeKey(k)
		if err := a.set(key, v, force); err != nil {
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
func (a *Attribute) set(key, value string, force bool) error {
	currentValue := git.Config.Find(key)
	if force || shouldReset(currentValue) {
		git.Config.UnsetGlobal(key)
		git.Config.SetGlobal(key, value)

		return nil
	} else if currentValue != value {
		return fmt.Errorf("The %s attribute should be \"%s\" but is \"%s\"",
			key, value, currentValue)
	}

	return nil
}

// Uninstall removes all properties in the path of this property.
func (a *Attribute) Uninstall() {
	git.Config.UnsetGlobalSection(a.Section)
}

// shouldReset determines whether or not a value is resettable given its current
// value on the system. If the value is empty (length = 0), then it will pass.
// Otherwise, it will pass if the below regex matches.
func shouldReset(value string) bool {
	if len(value) == 0 {
		return true
	}

	return valueRegexp.MatchString(value)
}
