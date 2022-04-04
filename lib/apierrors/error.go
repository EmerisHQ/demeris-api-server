package apierrors

import (
	"fmt"
)

type Error struct {
	Namespace     string
	StatusCode    int
	LowLevelError error
	Cause         string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Namespace, e.LowLevelError.Error())
}

func (e Error) Unwrap() error {
	return e.LowLevelError
}

func New(namespace string, cause error, statusCode int) Error {
	return Error{
		StatusCode:    statusCode,
		Namespace:     namespace,
		LowLevelError: cause,
		Cause:         cause.Error(),
	}
}
