package main

import (
	"fmt"
	idempotency "go-helloworld/http/basic/handler/user/get"
	"go-helloworld/http/basic/middleware/httperror"
	"go-helloworld/http/basic/middleware/recover"
	"log"
	"net/http"
	"strings"
	"time"
)

// helloHandler is a regular function that matches the signature (w http.ResponseWriter, r *http.Request).
// It will be wrapped into http.HandlerFunc by the multiplexer (or explicitly),
// allowing it to be used as an http.Handler.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Hello, %s \n", name)
}

func echoHandler(w http.ResponseWriter, r *http.Request) error {
	echo := r.URL.Query().Get("echo")
	if strings.TrimSpace(echo) == "" {
		return httperror.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("empty echo param"),
		)
	}

	fmt.Fprintf(w, "echo %s\n", echo)
	return nil
}

// userHandler handles requests like /user/123 and extracts the ID from the path.
// Note: http.ServeMux does not support parameters, so we manually parse URL.Path here.
func userHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) <= 3 {
		id := parts[2]
		fmt.Fprintf(w, "User ID %s \n", id)
	} else {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
	}
}

// AuthHandler is a custom struct that implements http.Handler interface
// by defining a ServeHTTP method.
type AuthHandler struct {
	secret string
}

func (a *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Auth you with my custom auth handler and secret %s! \n", a.secret)
}

// logRequestsMiddleware is a standard middleware pattern.
// It takes an http.Handler, wraps it, and returns a new http.Handler that logs requests.
func logRequestsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Create a new HTTP request multiplexer (router).
	// It matches the incoming request path against the registered routes.
	mux := http.NewServeMux()

	mux.Handle("/echo", recover.RecoverMiddleware(httperror.HTTPErrorMiddleware(echoHandler)))

	// Register a handler function for "/user/".
	// The function is automatically wrapped into http.HandlerFunc,
	// making it compatible with http.Handler.
	mux.HandleFunc("/user/", userHandler)

	idempotentUserHandler := idempotency.NewIdempotentHandler(10*time.Second, 10*time.Minute)

	mux.Handle("/idempotency/user", idempotentUserHandler)

	// Register a handler for "/hello", wrapped with logging middleware.
	// We explicitly wrap helloHandler with http.HandlerFunc to make it a http.Handler,
	// since logRequestsMiddleware expects a handler, not just a function.
	mux.Handle("/hello", logRequestsMiddleware(http.HandlerFunc(helloHandler)))

	// Register a custom handler that is a struct implementing http.Handler.
	customHandler := &AuthHandler{secret: "foo"}
	mux.Handle("/auth", customHandler)

	// Create a handler composed of built-in components:
	// StripPrefix removes the "/static/" prefix from the request URL,
	// and FileServer serves static files from the "./public" directory.
	// Example: "/static/images/logo.png" becomes "./public/images/logo.png"
	staticFilesHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./public")))
	mux.Handle("/static/", staticFilesHandler)

	// Create and configure the HTTP server.
	// We use mux as the root handler — it will receive all requests and dispatch accordingly.
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Server is running on http://localhost%s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
