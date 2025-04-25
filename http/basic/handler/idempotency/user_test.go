package idempotency

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IdempotentUserHandler_CacheMissThenHit(t *testing.T) {
	handler := NewIdempotentHandler(2*time.Second, 5*time.Minute)

	server := httptest.NewServer(handler)
	defer server.Close()

	client := http.Client{}

	req1, _ := http.NewRequest("GET", server.URL+"?id=1", nil)
	req1.Header.Set("Idempotency-Key", "key-123")
	resp1, err := client.Do(req1)
	require.NoError(t, err)
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()
	assert.Equal(t, "MISS", resp1.Header.Get("X-Cache"))

	req2, _ := http.NewRequest("GET", server.URL+"?id=1", nil)
	req2.Header.Set("Idempotency-Key", "key-123")
	resp2, err := client.Do(req2)
	require.NoError(t, err)
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	assert.Equal(t, "HIT", resp2.Header.Get("X-Cache"))
	assert.True(t, bytes.Equal(body1, body2))
}

func Test_IdempotentUserHandler_ParallelRequests(t *testing.T) {
	handler := NewIdempotentHandler(2*time.Second, 5*time.Minute)

	server := httptest.NewServer(handler)
	defer server.Close()

	var wg sync.WaitGroup
	const parallel = 50

	results := make([][]byte, parallel)
	errors := make([]error, parallel)
	client := http.Client{}

	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req, _ := http.NewRequest("GET", server.URL+"?id=1", nil)
			req.Header.Set("Idempotency-Key", "shared-key")
			resp, err := client.Do(req)
			if err != nil {
				errors[i] = err
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			results[i] = body
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
