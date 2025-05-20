package httperror

type HTTPError struct {
	Err  error
	Code int
}

func (e *HTTPError) Error() string {
	return e.Err.Error()
}

func NewHTTPError(code int, err error) *HTTPError {
	return &HTTPError{
		Code: code,
		Err:  err,
	}
}
