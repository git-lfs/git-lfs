package lfs

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

	fatal := newFatalError(err)

	if !IsFatalError(fatal) {
		t.Error("expected error to be fatal")
	}
}

func TestBehaviorWraps(t *testing.T) {
	err := errors.New("Go error")

	fatal := newFatalError(err)
	ni := newNotImplementedError(fatal)

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

	ErrorSetContext(err, "foo", "bar")

	v := ErrorGetContext(err, "foo")
	if v == "bar" {
		t.Error("expected empty context on go error")
	}
}

func TestContextOnWrappedErrors(t *testing.T) {
	err := newFatalError(errors.New("Go error"))

	ErrorSetContext(err, "foo", "bar")

	if v := ErrorGetContext(err, "foo"); v != "bar" {
		t.Error("expected to be able to use context on wrapped errors")
	}

	ctxt := ErrorContext(err)
	if ctxt["foo"] != "bar" {
		t.Error("expected to get the context of an error")
	}

	ErrorDelContext(err, "foo")

	if v := ErrorGetContext(err, "foo"); v == "bar" {
		t.Errorf("expected to delete from error context")
	}
}

func TestStack(t *testing.T) {
	s := ErrorStack(errors.New("Go error"))
	if len(s) > 0 {
		t.Error("expected to get no stack from a Go error")
	}

	s = ErrorStack(newFatalError(errors.New("Go error")))
	if len(s) == 0 {
		t.Error("expected to get a stack from a wrapped error")
	}
}
