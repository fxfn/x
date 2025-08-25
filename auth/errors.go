package auth

type InvalidClientError struct {
	error

	message string
}

func (e *InvalidClientError) Error() string {
	return e.message
}

func (e *InvalidClientError) Unwrap() error {
	return e.error
}

type InvalidRequest struct {
	error

	message string
}

func (e *InvalidRequest) Error() string {
	return e.message
}

func (e *InvalidRequest) Unwrap() error {
	return e.error
}
