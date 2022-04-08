package apierrors

import (
	"fmt"
)

type Error struct {
	Namespace  string
	StatusCode int
	Cause      string

	InternalCause    error
	LogKeysAndValues []any
}

func (e *Error) Error() string {
	c := e.Cause
	if e.InternalCause != nil {
		c = e.InternalCause.Error()
	}

	return fmt.Sprintf("%s: %s", e.Namespace, c)
}

func (e *Error) Unwrap() error {
	return e.InternalCause
}

func (e *Error) WithLogContext(internalCause error, keysAndValues ...any) *Error {
	e.InternalCause = internalCause
	e.LogKeysAndValues = keysAndValues
	return e
}

func New(namespace string, cause string, statusCode int) *Error {
	return &Error{
		StatusCode: statusCode,
		Namespace:  namespace,
		Cause:      cause,
	}
}
