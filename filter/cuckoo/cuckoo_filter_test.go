package filter

import (
	"fmt"
	"math/rand"
	"testing"
)

// TestInsertAndContains verifies that inserted keys are always found.
func TestInsertAndContains(t *testing.T) {
	rand.Seed(42)

	cf := NewCuckooFilter(256, 4, 500)

	keys := [][]byte{
		[]byte("apple"),
		[]byte("banana"),
		[]byte("orange"),
		[]byte("grape"),
		[]byte("watermelon"),
	}

	for _, k := range keys {
		if !cf.Insert(k) {
			t.Fatalf("failed to insert key: %s", k)
		}
	}

	for _, k := range keys {
		if !cf.Contains(k) {
			t.Fatalf("filter does not contain inserted key: %s", k)
		}
	}
}

// TestNoFalseNegatives ensures that filter never returns false for inserted keys.
func TestNoFalseNegatives(t *testing.T) {
	rand.Seed(1)

	cf := NewCuckooFilter(512, 4, 500)

	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("key-%d", i))

		if !cf.Insert(key) {
			t.Fatalf("insert failed for key %s", key)
		}

		if !cf.Contains(key) {
			t.Fatalf("false negative detected for key %s", key)
		}
	}
}

// TestFalsePositiveRate checks that false positives exist but are reasonably low.
func TestFalsePositiveRate(t *testing.T) {
	rand.Seed(99)

	cf := NewCuckooFilter(1024, 4, 500)

	inserted := 2000
	queries := 5000
	falsePositives := 0

	for i := 0; i < inserted; i++ {
		key := []byte(fmt.Sprintf("inserted-%d", i))
		if !cf.Insert(key) {
			t.Fatalf("insert failed at %d", i)
		}
	}

	for i := 0; i < queries; i++ {
		key := []byte(fmt.Sprintf("not-inserted-%d", i))
		if cf.Contains(key) {
			falsePositives++
		}
	}

	fpRate := float64(falsePositives) / float64(queries)

	// Fingerprint = 8 bit, bucketSize = 4 → expected FP < ~3%
	if fpRate > 0.05 {
		t.Fatalf("false positive rate too high: %.4f", fpRate)
	}
}

// TestInsertOverflow verifies behavior when filter is overfilled.
func TestInsertOverflow(t *testing.T) {
	rand.Seed(7)

	cf := NewCuckooFilter(64, 2, 50)

	failures := 0
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("overflow-%d", i))
		if !cf.Insert(key) {
			failures++
		}
	}

	if failures == 0 {
		t.Fatal("expected some insert failures, got none")
	}
}
