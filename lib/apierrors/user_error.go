package apierrors

// UserFacingError is a struct that can be returned as an API response in case
// of errors. The cause should not leak any implementation detail to the user
// but can be informative of what went wrong.
type UserFacingError struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Cause     string `json:"cause"`
}

func NewUserFacingError(id string, e *Error) UserFacingError {
	return UserFacingError{
		ID:        id,
		Namespace: e.Namespace,
		Cause:     e.Cause,
	}
}
