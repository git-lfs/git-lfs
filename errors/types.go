package errors

import (
	goerrors "errors"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/pkg/errors"
)

// IsFatalError indicates that the error is fatal and the process should exit
// immediately after handling the error.
func IsFatalError(err error) bool {
	if e, ok := err.(interface {
		Fatal() bool
	}); ok {
		return e.Fatal()
	}
	if parent := parentOf(err); parent != nil {
		return IsFatalError(parent)
	}
	return false
}

// IsNotImplementedError indicates the client attempted to use a feature the
// server has not implemented (e.g. the batch endpoint).
func IsNotImplementedError(err error) bool {
	if e, ok := err.(interface {
		NotImplemented() bool
	}); ok {
		return e.NotImplemented()
	}
	if parent := parentOf(err); parent != nil {
		return IsNotImplementedError(parent)
	}
	return false
}

// IsAuthError indicates the client provided a request with invalid or no
// authentication credentials when credentials are required (e.g. HTTP 401).
func IsAuthError(err error) bool {
	if e, ok := err.(interface {
		AuthError() bool
	}); ok {
		return e.AuthError()
	}
	if parent := parentOf(err); parent != nil {
		return IsAuthError(parent)
	}
	return false
}

// IsSmudgeError indicates an error while smudging a files.
func IsSmudgeError(err error) bool {
	if e, ok := err.(interface {
		SmudgeError() bool
	}); ok {
		return e.SmudgeError()
	}
	if parent := parentOf(err); parent != nil {
		return IsSmudgeError(parent)
	}
	return false
}

// IsCleanPointerError indicates an error while cleaning a file.
func IsCleanPointerError(err error) bool {
	if e, ok := err.(interface {
		CleanPointerError() bool
	}); ok {
		return e.CleanPointerError()
	}
	if parent := parentOf(err); parent != nil {
		return IsCleanPointerError(parent)
	}
	return false
}

// IsNotAPointerError indicates the parsed data is not an LFS pointer.
func IsNotAPointerError(err error) bool {
	if e, ok := err.(interface {
		NotAPointerError() bool
	}); ok {
		return e.NotAPointerError()
	}
	if parent := parentOf(err); parent != nil {
		return IsNotAPointerError(parent)
	}
	return false
}

// IsNotAPointerError indicates the parsed data is not an LFS pointer.
func IsPointerScanError(err error) bool {
	if e, ok := err.(interface {
		PointerScanError() bool
	}); ok {
		return e.PointerScanError()
	}
	if parent := parentOf(err); parent != nil {
		return IsPointerScanError(parent)
	}
	return false
}

// IsBadPointerKeyError indicates that the parsed data has an invalid key.
func IsBadPointerKeyError(err error) bool {
	if e, ok := err.(interface {
		BadPointerKeyError() bool
	}); ok {
		return e.BadPointerKeyError()
	}

	if parent := parentOf(err); parent != nil {
		return IsBadPointerKeyError(parent)
	}
	return false
}

// IsProtocolError indicates that the SSH pkt-line protocol data is invalid.
func IsProtocolError(err error) bool {
	if e, ok := err.(interface {
		ProtocolError() bool
	}); ok {
		return e.ProtocolError()
	}

	if parent := parentOf(err); parent != nil {
		return IsProtocolError(parent)
	}
	return false
}

// If an error is abad pointer error of any type, returns NotAPointerError
func StandardizeBadPointerError(err error) error {
	if IsBadPointerKeyError(err) {
		badErr := err.(badPointerKeyError)
		if badErr.Expected == "version" {
			return NewNotAPointerError(err)
		}
	}
	return err
}

// IsDownloadDeclinedError indicates that the smudge operation should not download.
// TODO: I don't really like using errors to control that flow, it should be refactored.
func IsDownloadDeclinedError(err error) bool {
	if e, ok := err.(interface {
		DownloadDeclinedError() bool
	}); ok {
		return e.DownloadDeclinedError()
	}
	if parent := parentOf(err); parent != nil {
		return IsDownloadDeclinedError(parent)
	}
	return false
}

// IsDownloadDeclinedError indicates that the upload operation failed because of
// an HTTP 422 response code.
func IsUnprocessableEntityError(err error) bool {
	if e, ok := err.(interface {
		UnprocessableEntityError() bool
	}); ok {
		return e.UnprocessableEntityError()
	}
	if parent := parentOf(err); parent != nil {
		return IsUnprocessableEntityError(parent)
	}
	return false
}

// IsRetriableError indicates the low level transfer had an error but the
// caller may retry the operation.
func IsRetriableError(err error) bool {
	if e, ok := err.(interface {
		RetriableError() bool
	}); ok {
		return e.RetriableError()
	}
	if cause, ok := Cause(err).(*url.Error); ok {
		return cause.Temporary() || cause.Timeout()
	}
	if parent := parentOf(err); parent != nil {
		return IsRetriableError(parent)
	}
	return false
}

func IsRetriableLaterError(err error) (time.Time, bool) {
	if e, ok := err.(interface {
		RetriableLaterError() (time.Time, bool)
	}); ok {
		return e.RetriableLaterError()
	}
	if parent := parentOf(err); parent != nil {
		return IsRetriableLaterError(parent)
	}
	return time.Time{}, false
}

type errorWithCause interface {
	Cause() error
	StackTrace() errors.StackTrace
	error
	fmt.Formatter
}

// wrappedError is the base error wrapper. It provides a Message string, a
// stack, and a context map around a regular Go error.
type wrappedError struct {
	errorWithCause
	context map[string]interface{}
}

// newWrappedError creates a wrappedError.
func newWrappedError(err error, message string) *wrappedError {
	if err == nil {
		err = errors.New(tr.Tr.Get("Error"))
	}

	var errWithCause errorWithCause

	if len(message) > 0 {
		errWithCause = errors.Wrap(err, message).(errorWithCause)
	} else if ewc, ok := err.(errorWithCause); ok {
		errWithCause = ewc
	} else {
		errWithCause = errors.Wrap(err, "LFS").(errorWithCause)
	}

	return &wrappedError{
		context:        make(map[string]interface{}),
		errorWithCause: errWithCause,
	}
}

// Set sets the value for the key in the context.
func (e wrappedError) Set(key string, val interface{}) {
	e.context[key] = val
}

// Get gets the value for a key in the context.
func (e wrappedError) Get(key string) interface{} {
	return e.context[key]
}

// Del removes a key from the context.
func (e wrappedError) Del(key string) {
	delete(e.context, key)
}

// Context returns the underlying context.
func (e wrappedError) Context() map[string]interface{} {
	return e.context
}

// Definitions for IsFatalError()

type fatalError struct {
	*wrappedError
}

func (e fatalError) Fatal() bool {
	return true
}

func NewFatalError(err error) error {
	return fatalError{newWrappedError(err, tr.Tr.Get("Fatal error"))}
}

// Definitions for IsNotImplementedError()

type notImplementedError struct {
	*wrappedError
}

func (e notImplementedError) NotImplemented() bool {
	return true
}

func NewNotImplementedError(err error) error {
	return notImplementedError{newWrappedError(err, tr.Tr.Get("Not implemented"))}
}

// Definitions for IsAuthError()

type authError struct {
	*wrappedError
}

func (e authError) AuthError() bool {
	return true
}

func NewAuthError(err error) error {
	return authError{newWrappedError(err, tr.Tr.Get("Authentication required"))}
}

// Definitions for IsSmudgeError()

type smudgeError struct {
	*wrappedError
}

func (e smudgeError) SmudgeError() bool {
	return true
}

func NewSmudgeError(err error, oid, filename string) error {
	e := smudgeError{newWrappedError(err, tr.Tr.Get("Smudge error"))}
	SetContext(e, "OID", oid)
	SetContext(e, "FileName", filename)
	return e
}

// Definitions for IsCleanPointerError()

type cleanPointerError struct {
	*wrappedError
}

func (e cleanPointerError) CleanPointerError() bool {
	return true
}

func NewCleanPointerError(pointer interface{}, bytes []byte) error {
	err := New(tr.Tr.Get("pointer error"))
	e := cleanPointerError{newWrappedError(err, "clean")}
	SetContext(e, "pointer", pointer)
	SetContext(e, "bytes", bytes)
	return e
}

// Definitions for IsNotAPointerError()

type notAPointerError struct {
	*wrappedError
}

func (e notAPointerError) NotAPointerError() bool {
	return true
}

func NewNotAPointerError(err error) error {
	return notAPointerError{newWrappedError(err, tr.Tr.Get("Pointer file error"))}
}

// Definitions for IsPointerScanError()

type PointerScanError struct {
	treeishOid string
	path       string
	*wrappedError
}

func (e PointerScanError) PointerScanError() bool {
	return true
}

func (e PointerScanError) OID() string {
	return e.treeishOid
}

func (e PointerScanError) Path() string {
	return e.path
}

func NewPointerScanError(err error, treeishOid, path string) error {
	return PointerScanError{treeishOid, path, newWrappedError(err, tr.Tr.Get("Pointer error"))}
}

type badPointerKeyError struct {
	Expected string
	Actual   string

	*wrappedError
}

func (e badPointerKeyError) BadPointerKeyError() bool {
	return true
}

func NewBadPointerKeyError(expected, actual string) error {
	err := Errorf(tr.Tr.Get("Expected key %s, got %s", expected, actual))
	return badPointerKeyError{expected, actual, newWrappedError(err, tr.Tr.Get("pointer parsing"))}
}

// Definitions for IsDownloadDeclinedError()

type downloadDeclinedError struct {
	*wrappedError
}

func (e downloadDeclinedError) DownloadDeclinedError() bool {
	return true
}

func NewDownloadDeclinedError(err error, msg string) error {
	return downloadDeclinedError{newWrappedError(err, msg)}
}

// Definitions for IsRetriableLaterError()

type retriableLaterError struct {
	*wrappedError
	timeAvailable time.Time
}

func NewRetriableLaterError(err error, header string) error {
	secs, err := strconv.Atoi(header)
	if err == nil {
		return retriableLaterError{
			wrappedError:  newWrappedError(err, ""),
			timeAvailable: time.Now().Add(time.Duration(secs) * time.Second),
		}
	}

	time, err := time.Parse(time.RFC1123, header)
	if err == nil {
		return retriableLaterError{
			wrappedError:  newWrappedError(err, ""),
			timeAvailable: time,
		}
	}

	// We could not return a successful error from the Retry-After header.
	return nil
}

func (e retriableLaterError) RetriableLaterError() (time.Time, bool) {
	return e.timeAvailable, true
}

// Definitions for IsUnprocessableEntityError()

type unprocessableEntityError struct {
	*wrappedError
}

func (e unprocessableEntityError) UnprocessableEntityError() bool {
	return true
}

func NewUnprocessableEntityError(err error) error {
	return unprocessableEntityError{newWrappedError(err, "")}
}

// Definitions for IsRetriableError()

type retriableError struct {
	*wrappedError
}

func (e retriableError) RetriableError() bool {
	return true
}

func NewRetriableError(err error) error {
	return retriableError{newWrappedError(err, "")}
}

// Definitions for IsProtocolError()

type protocolError struct {
	*wrappedError
}

func (e protocolError) ProtocolError() bool {
	return true
}

func NewProtocolError(message string, err error) error {
	return protocolError{newWrappedError(err, message)}
}

func parentOf(err error) error {
	type causer interface {
		Cause() error
	}

	if c, ok := err.(causer); ok {
		if innerC, innerOk := c.Cause().(causer); innerOk {
			return innerC.Cause()
		}
	}

	return nil
}

func ExitStatus(err error) int {
	var eerr *exec.ExitError
	if goerrors.As(err, &eerr) {
		ws, ok := eerr.ProcessState.Sys().(syscall.WaitStatus)
		if ok {
			return ws.ExitStatus()
		}
	}
	return -1
}
