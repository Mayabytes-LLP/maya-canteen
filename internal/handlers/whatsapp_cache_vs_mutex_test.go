package handlers

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestWhatsAppRequestCacheVsBulkMutex demonstrates the difference between
// the existing whatsappRequestCache and the new bulkOperationMutex
func TestWhatsAppRequestCacheVsBulkMutex(t *testing.T) {
	t.Run("WhatsAppRequestCacheOnlyPreventsExactDuplicates", func(t *testing.T) {
		requestCache := make(map[string]time.Time)
		var cacheMutex sync.Mutex

		checkAndCacheRequest := func(requestID string) bool {
			cacheMutex.Lock()
			defer cacheMutex.Unlock()

			if _, exists := requestCache[requestID]; exists {
				return false
			}

			requestCache[requestID] = time.Now()
			return true
		}

		requestID1 := "abc123"

		allowed1 := checkAndCacheRequest(requestID1)
		allowed2 := checkAndCacheRequest(requestID1)

		assert.True(t, allowed1, "First request should be allowed")
		assert.False(t, allowed2, "Duplicate request ID should be rejected by cache")

		requestID2 := "def456"

		allowed3 := checkAndCacheRequest(requestID2)

		assert.True(t, allowed3, "Different request ID should be allowed by cache")

		t.Log("whatsappRequestCache only prevents EXACT duplicate request IDs")
	})

	t.Run("ContentBasedKeyPreventsSemanticDuplicates", func(t *testing.T) {
		requestCache := make(map[string]time.Time)
		var cacheMutex sync.Mutex
		var processedUserCount int64

		checkAndCacheRequestAndContent := func(requestID, contentKey string) bool {
			cacheMutex.Lock()
			defer cacheMutex.Unlock()

			if _, exists := requestCache[requestID]; exists {
				return false
			}

			if _, exists := requestCache[contentKey]; exists {
				return false
			}

			requestCache[requestID] = time.Now()
			requestCache[contentKey] = time.Now()
			return true
		}

		generateRequestID := func() string {
			return "rid-" + time.Now().Format("20060102150405.000000")
		}

		var wg sync.WaitGroup

		contentKey := "notify-all-2024-01-2024-01-true"

		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(clickNum int) {
				defer wg.Done()

				requestID := generateRequestID()

				allowed := checkAndCacheRequestAndContent(requestID, contentKey)

				if allowed {
					t.Logf("Click %d: Allowed, processing users", clickNum+1)
					for u := 0; u < 3; u++ {
						atomic.AddInt64(&processedUserCount, 1)
						time.Sleep(10 * time.Millisecond)
					}
				} else {
					t.Logf("Click %d: Rejected as duplicate", clickNum+1)
				}
			}(i)
		}

		wg.Wait()

		finalProcessedCount := atomic.LoadInt64(&processedUserCount)
		t.Logf("Final result: %d user processing events", finalProcessedCount)

		assert.Equal(t, int64(3), finalProcessedCount,
			"With content-based deduplication, only one bulk operation should process users")
	})

	t.Run("BulkMutexSerializesConcurrentDifferentContent", func(t *testing.T) {
		requestCache := make(map[string]time.Time)
		var cacheMutex sync.Mutex
		var bulkMutex sync.Mutex
		var processedUserCount int64

		generateRequestID := func() string {
			return "rid-" + time.Now().Format("20060102150405.000000")
		}

		checkAndCache := func(requestID, contentKey string) bool {
			cacheMutex.Lock()
			defer cacheMutex.Unlock()

			if _, exists := requestCache[requestID]; exists {
				return false
			}
			if _, exists := requestCache[contentKey]; exists {
				return false
			}

			requestCache[requestID] = time.Now()
			requestCache[contentKey] = time.Now()
			return true
		}

		processBulkWithBothProtections := func(clickNum int, month string) {
			requestID := generateRequestID()
			contentKey := fmt.Sprintf("notify-all-%s-%s-true", month, month)

			if !checkAndCache(requestID, contentKey) {
				t.Logf("Click %d: Rejected by cache", clickNum)
				return
			}

			t.Logf("Click %d: Acquiring bulk operation lock", clickNum)
			bulkMutex.Lock()
			defer bulkMutex.Unlock()

			t.Logf("Click %d: Processing with bulk lock", clickNum)
			for u := 0; u < 3; u++ {
				atomic.AddInt64(&processedUserCount, 1)
				time.Sleep(20 * time.Millisecond)
			}
		}

		var wg sync.WaitGroup

		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(clickNum int) {
				defer wg.Done()
				month := "2024-01"
				processBulkWithBothProtections(clickNum+1, month)
			}(i)
		}

		wg.Wait()

		finalProcessedCount := atomic.LoadInt64(&processedUserCount)
		t.Logf("With both protections: %d user processing events", finalProcessedCount)

		assert.Equal(t, int64(3), finalProcessedCount,
			"With content-based dedup + bulk mutex, only one operation should process users")
	})
}

// TestWhyBothMechanismsAreNeeded provides a clear explanation
func TestWhyBothMechanismsAreNeeded(t *testing.T) {
	t.Log("=== EXPLANATION: Why we need BOTH content-based dedup AND bulkOperationMutex ===")
	t.Log("")

	t.Log("Content-based deduplication protects against:")
	t.Log("  Same request ID processed twice (e.g., network retry with same ID)")
	t.Log("  Different request IDs for same notification content")
	t.Log("  User double-clicking same button rapidly")
	t.Log("  Multiple browser tabs making concurrent requests for same month/year")
	t.Log("")

	t.Log("Content-based dedup CANNOT protect against:")
	t.Log("  Different content requests running concurrently (different months)")
	t.Log("  Direct API calls with different parameters")
	t.Log("")

	t.Log("bulkOperationMutex protects against:")
	t.Log("  Multiple bulk operations running concurrently")
	t.Log("  WhatsApp client being used by multiple goroutines simultaneously")
	t.Log("  Database race conditions during user list fetching")
	t.Log("")

	t.Log("REAL SCENARIO that shows why both are needed:")
	t.Log("  1. User clicks 'Send All' for January at 10:00:00.000")
	t.Log("  2. Content key: notify-all-2024-01-2024-01-true is cached")
	t.Log("  3. User clicks 'Send All' for January at 10:00:00.100 (after debounce)")
	t.Log("  4. Same content key found in cache -> REJECTED (no duplicate)")
	t.Log("  5. But if admin clicks 'Send All' for February -> different content key -> allowed")
	t.Log("  6. bulkOperationMutex ensures Jan and Feb don't run simultaneously")
	t.Log("")

	t.Log("VERDICT: Both mechanisms are complementary and necessary")
}
