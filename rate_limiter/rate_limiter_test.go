package rate_limiter

import (
	"bytes"
	"io"
	"log"
	"testing"
	"time"
)

type multiWriter struct {
	writers []io.Writer
}

func (m *multiWriter) Write(p []byte) (n int, err error) {
	for _, writer := range m.writers {
		writer.Write(p)
	}

	return len(p), nil
}

func sendMessagesWithRateLimit(messages []string, tickInterval time.Duration, maxRequests int) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop() // Not necessary

	messageIndex := 0

	for range ticker.C {
		log.Printf("Tick received at: %s", time.Now())
		if messageIndex >= len(messages) {
			log.Println("All messages sent, stopping.")
			break
		}

		requestsCounts := 0

		for requestsCounts < maxRequests && messageIndex < len(messages) {
			log.Printf("Sending message %s\n", messages[messageIndex])
			messageIndex++
			requestsCounts++
		}
	}
}

func TestSendMessagesWithRateLimit(t *testing.T) {
	tests := []struct {
		name             string
		messages         []string
		tickInterval     time.Duration
		maxRequests      int
		expectedMessages []string
	}{
		{
			name:             "Test 3 messages per second",
			messages:         []string{"Hello", "World", "Go", "Rocks", "Learn", "Ticker", "Example"},
			tickInterval:     time.Second,
			maxRequests:      3,
			expectedMessages: []string{"Hello", "World", "Go", "Rocks", "Learn", "Ticker", "Example"},
		},
		{
			name:             "Test 1 message per second",
			messages:         []string{"Message 1", "Message 2", "Message 3"},
			tickInterval:     time.Second,
			maxRequests:      1,
			expectedMessages: []string{"Message 1", "Message 2", "Message 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			multi := &multiWriter{
				writers: []io.Writer{
					&buf,
					log.Writer(),
				},
			}

			log.SetOutput(multi)

			sendMessagesWithRateLimit(tt.messages, tt.tickInterval, tt.maxRequests)

			for _, expected := range tt.expectedMessages {
				if !bytes.Contains(buf.Bytes(), []byte(expected)) {
					t.Errorf("Expected message %q, but it wasn't found", expected)
				}
			}
		})
	}
}
