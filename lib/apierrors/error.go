package apierrors

import (
	"fmt"
)

type Error struct {
	Namespace  string
	StatusCode int
	Cause      string

	InternalCause    string
	LogKeysAndValues []any
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Namespace, e.Cause)
}

func (e *Error) WithLogContext(internalCause string, keysAndValues ...any) *Error {
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
