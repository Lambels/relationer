package internal

import (
	"errors"
	"fmt"
	"net/http"
)

// ECode represents an error code in the system.
type ECode uint8

// Error codes which map good to http errors.
const (
	ECONFLICT ECode = iota + 1
	EINTERNAL
	EINVALID
	ENOTFOUND
)

var codes = map[ECode]int{
	ECONFLICT: http.StatusConflict,
	EINVALID:  http.StatusBadRequest,
	ENOTFOUND: http.StatusNotFound,
	EINTERNAL: http.StatusInternalServerError,
}

// Error represents an internal error which implements the error interface.
type Error struct {
	// the origin of the current error.
	origin error

	// message is the human readable message.
	message string

	// code is a machine readable code.
	code ECode
}

// WrapError wraps the origin error in the Error type with the formated new error.
func WrapError(origin error, code ECode, format string, a ...interface{}) error {
	return &Error{
		origin:  origin,
		message: fmt.Sprintf(format, a...),
		code:    code,
	}
}

func WrapErrorNil(origin error, code ECode, format string, a ...interface{}) error {
	if origin == nil {
		return nil
	}
	return WrapError(origin, code, format, a...)
}

// Errorf formats a new error.
func Errorf(code ECode, formant string, a ...interface{}) error {
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

// ErrorCode is a helper function to get the code for any error.
//
// if error nil: returns 0
//
// if error not of type internal.Error: EINTERNAL
//
// if error of type internal.Error: code of error
func ErrorCode(err error) ECode {
	var e *Error
	if err == nil {
		return 0
	} else if errors.As(err, &e) {
		return e.Code()
	}
	return EINTERNAL
}

func StatusCodeFromECode(code ECode) int {
	v, ok := codes[code]
	if !ok {
		return http.StatusInternalServerError
	}
	return v
}

// ECodeFromStatusCode is a helper function to map a http status code to a error
//
// if code not found: returns EINTERNAL
func ECodeFromStatusCode(code int) ECode {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return EINTERNAL
}

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.origin
}

// Code returns the error code under the error.
func (e *Error) Code() ECode {
	return e.code
}
