package tq

import "github.com/git-lfs/git-lfs/v3/tr"

type MalformedObjectError struct {
	Name string
	Oid  string

	missing bool
}

func newObjectMissingError(name, oid string) error {
	return &MalformedObjectError{Name: name, Oid: oid, missing: true}
}

func newCorruptObjectError(name, oid string) error {
	return &MalformedObjectError{Name: name, Oid: oid, missing: false}
}

func (e MalformedObjectError) Missing() bool { return e.missing }

func (e MalformedObjectError) Corrupt() bool { return !e.Missing() }

func (e MalformedObjectError) Error() string {
	if e.Corrupt() {
		return tr.Tr.Get("corrupt object: %s (%s)", e.Name, e.Oid)
	}
	return tr.Tr.Get("missing object: %s (%s)", e.Name, e.Oid)
}
