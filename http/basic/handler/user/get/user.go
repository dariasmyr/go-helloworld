package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type HTTPError struct {
	Code int
	Err  error
}

func (e *HTTPError) Error() string {
	return e.Err.Error()
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

func (s *SingleFlight) Do(key string, fn func() (interface{}, error)) (val interface{}, err error) {
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
			log.Printf("panic recovered in SingleFlight for key=%s: %v", key, v)
			err = fmt.Errorf("panic: %v", v)
			newCall.err = err
			s.mu.Lock()
			delete(s.m, key)
			s.mu.Unlock()
		}
		newCall.wg.Done()
	}()

	newCall.val, newCall.err = fn()

	val = newCall.val
	err = newCall.err
	return
}

type CacheEntry struct {
	Value     []byte
	UpdatedAt time.Time
}

type IdempotentUserHandler struct {
	mu      sync.Mutex
	cache   map[string]CacheEntry
	ttl     time.Duration
	sf      *SingleFlight
	timeout time.Duration
}

func NewIdempotentHandler(timeout, ttl time.Duration) *IdempotentUserHandler {
	return &IdempotentUserHandler{
		timeout: timeout,
		cache:   make(map[string]CacheEntry),
		sf: &SingleFlight{
			m: make(map[string]*call),
		},
		ttl: ttl,
	}
}

func (i *IdempotentUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
					ctxBackground, ctxBackgroundCancel := context.WithTimeout(context.Background(), i.timeout)
					defer func() {
						log.Printf("background refresh done for key=%s", key)
						ctxBackgroundCancel()
					}()
					return task(ctxBackground, userID)
				})
				if err != nil {
					// we need not lose errors and at least keep it somewhere
					log.Printf("error updating user data in the background %v", err)
					return
				}

				data, ok := res.([]byte)
				if !ok {
					log.Printf("invalid type in SingleFlight result: expected []byte, got %T (key=%s)", res, key)
					return
				}

				// Save to cache
				i.mu.Lock()
				i.cache[key] = CacheEntry{
					Value:     data,
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
		return task(ctx, userID)
	})
	if err != nil {
		// we need not lose errors and at least keep it somewhere
		var httpErr *HTTPError
		if errors.As(err, &httpErr) {
			log.Printf("error processing user data with code %d: %v", httpErr.Code, httpErr.Err)
			http.Error(w, httpErr.Error(), httpErr.Code)
			return
		}
		// errors that have no wrapped HTTPErr error
		log.Printf("error processing user data %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, ok := res.([]byte)
	if !ok {
		errMsg := fmt.Sprintf("invalid type in SingleFlight result: expected []byte, got %T (key=%s)", res, key)
		log.Printf("error: %s", errMsg)
		http.Error(w, "interval server error", http.StatusInternalServerError)
		return
	}

	// Save to cache
	i.mu.Lock()
	i.cache[key] = CacheEntry{
		Value:     data,
		UpdatedAt: time.Now(),
	}
	i.mu.Unlock()

	// TODO Save to db
	w.Header().Set("X-Cache", "MISS")
	w.Write(data)
}

type userData struct {
	UserId string `json:"userId"`
	Data   string `json:"name"`
}

func task(ctx context.Context, userID string) ([]byte, error) {
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%s", userID)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", url, nil)
	if reqErr != nil {
		return nil, &HTTPError{
			Code: http.StatusInternalServerError,
			Err:  reqErr,
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &HTTPError{
			Code: http.StatusBadGateway, // as our handler work as a proxy to reach other resourse, so its bad gateway
			Err:  err,
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{
			Code: resp.StatusCode,
			Err:  fmt.Errorf("external API error: status %d", resp.StatusCode),
		}
	}

	var user userData
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, &HTTPError{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error decoding JSON: %w", err),
		}
	}

	// or use io.ReadAll(resp.Body) to get raw bytes

	fmt.Println(user.Data)

	return []byte(user.Data), nil
}

func taskWithRetry(ctx context.Context, userID string, maxRetries int) ([]byte, error) {
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%s", userID)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for i := 0; i <= maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request failed: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			var user userData
			if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
				return nil, fmt.Errorf("error decoding JSON: %w", err)
			}
			return []byte(user.Data), nil
		}

		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled")
			return nil, fmt.Errorf("context cancelled %w", ctx.Err())
		case <-time.After(time.Duration(i) * time.Second): // type converted i to time.Duration(i) to add incremental backoff retry
			if i == maxRetries {
				return nil, ctx.Err()
			}
			fmt.Println("Timeout")
		}
	}

	return nil, fmt.Errorf("max retries exceeded")
}

func generateIdempotencyKey(r *http.Request) string {
	hash := sha256.New()
	hash.Write([]byte(r.URL.String()))
	return hex.EncodeToString(hash.Sum(nil))
}

func taskWithPanic(ctx context.Context, userID string) ([]byte, error) {
	if rng.Intn(2)%2 == 0 {
		panic("panic while user handling")
	}

	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%s", userID)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, reqErr := http.NewRequestWithContext(ctx, "GET", url, nil)
	if reqErr != nil {
		return nil, &HTTPError{
			Code: http.StatusInternalServerError,
			Err:  reqErr,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, &HTTPError{
			Code: http.StatusInternalServerError,
			Err:  err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{
			Code: resp.StatusCode,
			Err:  fmt.Errorf("external API error: status %d", resp.StatusCode),
		}
	}

	var user userData

	decodeErr := json.NewDecoder(resp.Body).Decode(&user)
	if decodeErr != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Err: decodeErr}
	}

	return []byte(user.Data), nil
}
