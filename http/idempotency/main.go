package main

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type inFlight struct {
	wg     sync.WaitGroup
	result []byte
	err    error
}

type Middleware struct {
	redis *redis.Client
	ttl   time.Duration
	mu    sync.Mutex
	cache map[string]*inFlight
}

func New(redis *redis.Client, ttl time.Duration) *Middleware {
	return &Middleware{
		redis: redis,
		ttl:   ttl,
		cache: make(map[string]*inFlight),
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check Redis
		if cached, err := m.redis.Get(ctx, key).Bytes(); err == nil {
			w.Header().Set("X-Idempotent", "true")
			w.Write(cached)
			return
		}

		// In-memory singleflight
		m.mu.Lock()
		if call, exists := m.cache[key]; exists {
			m.mu.Unlock()
			call.wg.Wait()
			if call.err != nil {
				http.Error(w, call.err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(call.result)
			return
		}
		call := &inFlight{}
		call.wg.Add(1)
		m.cache[key] = call
		m.mu.Unlock()

		// Capture response
		rec := &responseRecorder{ResponseWriter: w, buf: &bytes.Buffer{}}
		next.ServeHTTP(rec, r)

		call.result = rec.buf.Bytes()
		call.err = nil // can error-capture by status
		call.wg.Done()

		m.mu.Lock()
		delete(m.cache, key)
		m.mu.Unlock()

		// Write to redis
		m.redis.Set(ctx, key, call.result, m.ttl)
		w.Header().Set("X-Idempotent", "false")
		w.Write(call.result)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.buf.Write(b)
	return r.ResponseWriter.Write(b)
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.publicapis.org/entries", nil)
	if err != nil {
		http.Error(w, "Failed to create external request", http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch external API", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	idemp := New(redisClient, 5*time.Minute)

	http.Handle("/submit", idemp.Handler(http.HandlerFunc(myHandler)))

}
