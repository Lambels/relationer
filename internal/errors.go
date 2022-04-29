package internal

import "fmt"

// ErrorCode represents an error code in the system.
type ErrorCode uint

// Error codes which map good to http errors.
const (
	ECONFLICT ErrorCode = iota
	EINTERNAL
	EINVALID
	ENOTFOUND
)

// Error represents an internal error which implements the error interface.
type Error struct {
	// the origin of the current error.
	origin error

	// message is the human readable message.
	message string

	// code is a machine readable code.
	code ErrorCode
}

// WrapError wraps the origin error in the Error type with the formated new error.
func WrapError(origin error, code ErrorCode, format string, a ...interface{}) error {
	return &Error{
		origin:  origin,
		message: fmt.Sprintf(format, a...),
		code:    code,
	}
}

// Errorf formats a new error.
func Errorf(code ErrorCode, formant string, a ...interface{}) error {
	return WrapError(nil, code, formant, a...)
}

// Error returns the message, when wrapped we return the wrapped error message with the current
// error message.
func (e *Error) Error() string {
	if e.origin != nil {
		return fmt.Sprintf("%v: %v", e.origin, e.message)
	}

	return e.message
}

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.origin
}

// Code returns the error code under the error.
func (e *Error) Code() ErrorCode {
	return e.code
}
