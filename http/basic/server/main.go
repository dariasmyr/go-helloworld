package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
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

// test handler with idempotency via singleflight processing
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type SingleFlight struct {
	mu sync.Mutex
	m  map[string]*call
}

func (s *SingleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	s.mu.Lock()
	if existingCall, exists := s.m[key]; exists {
		s.mu.Unlock()
		existingCall.wg.Wait()
		return existingCall.val, existingCall.err
	}

	newCall := &call{}
	newCall.wg.Add(1)
	s.m[key] = newCall
	s.mu.Unlock()
	defer func() {
		if v := recover(); v != nil {
			newCall.err = fmt.Errorf("panic: %v", v)
			s.mu.Lock()
			delete(s.m, key)
			s.mu.Unlock()
		}
	}()
	defer newCall.wg.Done()

	newCall.val, newCall.err = fn()

	return newCall.val, newCall.err
}

type CacheEntry struct {
	Value     []byte
	UpdatedAt time.Time
}

type IdempotentHandler struct {
	mu      sync.Mutex
	cache   map[string]CacheEntry
	ttl     time.Duration
	sf      *SingleFlight
	timeout time.Duration
}

func (i *IdempotentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), i.timeout)
	defer cancel()

	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	//look for idempotency key
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		key = generateIdempotencyKey(r)
	}

	i.mu.Lock()
	if data, ok := i.cache[key]; ok {
		if time.Since(data.UpdatedAt) > i.ttl {
			i.mu.Unlock()
			// update cache in the background with ctx background to avoid memory leak and control goroutine lifetime
			go func() {
				res, err := i.sf.Do(key, func() (interface{}, error) {
					return task(ctx, userID, i.timeout)
				})
				if err != nil {
					// we need not lose errors and at least keep it somewhere
					wErr := fmt.Errorf("error updating user data in the background %w", err)
					log.Printf(wErr.Error())
					return
				}

				// Save to cache
				i.mu.Lock()
				i.cache[key] = CacheEntry{
					Value:     res.([]byte),
					UpdatedAt: time.Now(),
				}
				i.mu.Unlock()

				// TODO Save to db
			}()
			// return cache that will be soon updated
			w.Header().Set("X-Cache", "STALE")
			w.Write(data.Value)
			return
		}
		w.Header().Set("X-Cache", "HIT")
		w.Write(data.Value)
		return
	}
	i.mu.Unlock()

	res, err := i.sf.Do(key, func() (interface{}, error) {
		return task(ctx, userID, i.timeout)
	})
	if err != nil {
		// we need not lose errors and at least keep it somewhere
		wErr := fmt.Errorf("error processing user data %w", err)
		log.Printf(wErr.Error())
		http.Error(w, fmt.Sprintf("error processing user data %w", err), http.StatusInternalServerError)
		return
	}

	// Save to cache
	i.mu.Lock()
	i.cache[key] = CacheEntry{
		Value:     res.([]byte),
		UpdatedAt: time.Now(),
	}
	i.mu.Unlock()

	// TODO Save to db
	w.Header().Set("X-Cache", "MISS")
	w.Write(res.([]byte))
}

type userData struct {
	UserId string `json:"userId"`
	Data   string `json:"data"`
}

func task(ctx context.Context, userID string, timeout time.Duration) ([]byte, error) {
	ctxBackground, ctxBackgroundCancel := context.WithTimeout(ctx, timeout)
	defer ctxBackgroundCancel()

	url := fmt.Sprintf("http://external-api.local/user?id=%s", userID)

	req, err := http.NewRequestWithContext(ctxBackground, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external API error: %w", err)
	}

	var user userData
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	// ir use io.ReadAll(resp.Body) to ger raw bytes

	return []byte(user.Data), nil
}

func generateIdempotencyKey(r *http.Request) string {
	hash := sha256.New()
	hash.Write([]byte(r.URL.String()))
	return hex.EncodeToString(hash.Sum(nil))
}

func main() {
	// Create a new HTTP request multiplexer (router).
	// It matches the incoming request path against the registered routes.
	mux := http.NewServeMux()

	// Register a handler function for "/user/".
	// The function is automatically wrapped into http.HandlerFunc,
	// making it compatible with http.Handler.
	mux.HandleFunc("/user/", userHandler)

	idempotentHandler := &IdempotentHandler{
		timeout: 10 * time.Second,
		cache:   make(map[string]CacheEntry),
		sf: &SingleFlight{
			m: make(map[string]*call),
		},
		ttl: 10 * time.Minute,
	}

	mux.Handle("/user", idempotentHandler)

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

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
