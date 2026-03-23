package errors

import (
	"fmt"

	je "github.com/juju/errors"
)

var (
	// Stack wraps je.ErrorStack
	Stack = je.ErrorStack
	// Trace wraps je.Trace
	Trace = je.Trace
	// Annotatef wraps je.Annotatef
	Annotatef = je.Annotatef
	// Errorf wraps je.Errorf
	Errorf = je.Errorf
	// Wrap wraps je.Wrap
	Wrap = je.Wrap
	// Cause wraps je.Cause
	Cause = je.Cause
)

// Error .
type Error struct {
	Code     int
	Synopsis string
	Detail   string
}

// NewError initiates a new Error.

func NewError(code int, synopsis, detail string) Error {
	return Error{
		Code:     code,
		Synopsis: synopsis,
		Detail:   detail,
	}
}

// Error implements error interface.
func (e *Error) Error() string {
	if len(e.Detail) > 0 {
		return fmt.Sprintf("%s: %s", e.Synopsis, e.Detail)
	}
	return e.Synopsis
}

// AnnotateDBErrorf annotates a DB error.
func AnnotateDBErrorf(err error, format string, args ...interface{}) Error {
	return annotateErrorf(err, DBErrorCode, DBErrorSynopsis, format, args...)
}

// AnnotateAppErrorf annotates an app error.
func AnnotateAppErrorf(err error, format string, args ...interface{}) Error {
	return annotateErrorf(err, AppErrorCode, AppErrorSynopsis, format, args...)
}

// AnnotateParameterErrorf annotates a parameter error.
func AnnotateParameterErrorf(err error, format string, args ...interface{}) Error {
	return annotateErrorf(err, ParameterErrorCode, ParameterErrorSynopsis, format, args...)
}

func annotateErrorf(err error, code int, synopsis, format string, args ...interface{}) Error {
	return NewError(
		code,
		synopsis,
		fmt.Sprintf("%s: %s", fmt.Sprintf(format, args...), err.Error()),
	)
}
