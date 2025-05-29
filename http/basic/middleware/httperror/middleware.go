package httperror

import (
	"errors"
	"net/http"
)

type HTTPHandlerWithError func(w http.ResponseWriter, r *http.Request) error

func HTTPErrorMiddleware(handler HTTPHandlerWithError) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			var httpError *HTTPError
			if errors.As(err, &httpError) {
				http.Error(w, httpError.Error(), httpError.Code)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return // necessary not to panic by http.WriteHeader re-writing
		}
	})
}
