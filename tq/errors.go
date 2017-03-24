package tq

import "fmt"

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
		return fmt.Sprintf("corrupt object: %s (%s)", e.Name, e.Oid)
	}
	return fmt.Sprintf("missing object: %s (%s)", e.Name, e.Oid)
}
