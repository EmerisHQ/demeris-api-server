package deps

import (
	"fmt"
)

type Error struct {
	ID            string `json:"id"`
	Namespace     string `json:"namespace"`
	StatusCode    int    `json:"-"`
	LowLevelError error  `json:"-"`
	Cause         string `json:"cause"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s @ %s: %s", e.ID, e.Namespace, e.LowLevelError.Error())
}

func (e Error) Unwrap() error {
	return e.LowLevelError
}

func NewError(namespace string, cause error, statusCode int) Error {
	return Error{
		StatusCode:    statusCode,
		Namespace:     namespace,
		LowLevelError: cause,
		Cause:         cause.Error(),
	}
}
