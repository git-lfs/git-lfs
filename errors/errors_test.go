package errors

import (
	"errors"
	"testing"
)

func TestChecksHandleGoErrors(t *testing.T) {
	err := errors.New("Go Error")

	if IsFatalError(err) {
		t.Error("go error should not be a fatal error")
	}
}

func TestCheckHandlesWrappedErrors(t *testing.T) {
	err := errors.New("Go error")

	fatal := NewFatalError(err)

	if !IsFatalError(fatal) {
		t.Error("expected error to be fatal")
	}
}

func TestBehaviorWraps(t *testing.T) {
	err := errors.New("Go error")

	fatal := NewFatalError(err)
	ni := NewNotImplementedError(fatal)

	if !IsNotImplementedError(ni) {
		t.Error("expected erro to be not implemeted")
	}

	if !IsFatalError(ni) {
		t.Error("expected wrapped error to also be fatal")
	}

	if IsNotImplementedError(fatal) {
		t.Error("expected fatal error to not be not implemented")
	}
}

func TestContextOnGoErrors(t *testing.T) {
	err := errors.New("Go error")

	SetContext(err, "foo", "bar")

	v := GetContext(err, "foo")
	if v == "bar" {
		t.Error("expected empty context on go error")
	}
}

func TestContextOnWrappedErrors(t *testing.T) {
	err := NewFatalError(errors.New("Go error"))

	SetContext(err, "foo", "bar")

	if v := GetContext(err, "foo"); v != "bar" {
		t.Error("expected to be able to use context on wrapped errors")
	}

	ctxt := Context(err)
	if ctxt["foo"] != "bar" {
		t.Error("expected to get the context of an error")
	}

	DelContext(err, "foo")

	if v := GetContext(err, "foo"); v == "bar" {
		t.Errorf("expected to delete from error context")
	}
}
