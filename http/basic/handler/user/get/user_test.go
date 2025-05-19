package idempotency

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_IdempotentUserHandler_CacheMissThenHit(t *testing.T) {
	handler := NewIdempotentHandler(2*time.Second, 5*time.Minute)

	server := httptest.NewServer(handler)
	defer server.Close()

	req1 := httptest.NewRequest("GET", "/?id=1", nil)
	req1.Header.Set("Idempotency-Key", "key-123")

	rec1 := httptest.NewRecorder()

	handler.ServeHTTP(rec1, req1)
	assert.Equal(t, "MISS", rec1.Header().Get("X-Cache"))

	req2 := httptest.NewRequest("GET", "/?id=1", nil)
	req2.Header.Set("Idempotency-Key", "key-123")
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	assert.Equal(t, "HIT", rec2.Header().Get("X-Cache"))
	assert.True(t, bytes.Equal(rec2.Body.Bytes(), rec2.Body.Bytes()))
}

func Test_IdempotentUserHandler_ParallelRequests(t *testing.T) {
	handler := NewIdempotentHandler(5*time.Second, 5*time.Minute)

	var wg sync.WaitGroup
	const parallel = 50

	results := make([][]byte, parallel)
	errors := make([]error, parallel)

	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/?id=1", nil)
			req.Header.Set("Idempotency-Key", "shared-key")

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			results[i] = []byte(rec.Body.String())
		}(i)
	}
	wg.Wait()

	for _, err := range errors {
		assert.NoError(t, err)
	}
	first := results[0]
	for i := 1; i < parallel; i++ {
		assert.Equal(t, first, results[i], "all results should be equal")
	}
}

func Test_IdempotentUserHandler_StaleCacheTriggersBackgroundRefresh(t *testing.T) {
	handler := NewIdempotentHandler(1*time.Second, 1*time.Millisecond)

	// prepopulate cache
	req := httptest.NewRequest("GET", "/?id=1", nil)
	req.Header.Set("Idempotency-Key", "k-stale")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, "MISS", rec.Header().Get("X-Cache"))

	// wait for TTL to expire
	time.Sleep(2 * time.Millisecond)

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	assert.Equal(t, "STALE", rec2.Header().Get("X-Cache"))
}

func Test_IdempotentUserHandler_InvalidId(t *testing.T) {
	handler := NewIdempotentHandler(1*time.Second, 1*time.Minute)

	req := httptest.NewRequest("GET", "/?id=", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Missing user ID")
}

func Test_IdempotentUserHandler_RandomPanic(t *testing.T) {
	handler := NewIdempotentHandler(1*time.Second, 1*time.Minute)

	req := httptest.NewRequest("GET", "/?id=1", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	body := rec.Body.String()
	fmt.Println("Response: ", body)

	assert.Contains(t, body, "panic while user handling")
}
