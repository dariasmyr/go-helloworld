package recover

import (
	"log"
	"net/http"
	"runtime/debug"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC]: %v\n%s", rec, debug.Stack())
				http.Error(w, "interval server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
