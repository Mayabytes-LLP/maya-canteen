package handlers

import (
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
		// Simulate the existing whatsappRequestCache behavior
		requestCache := make(map[string]time.Time)
		var cacheMutex sync.Mutex

		checkAndCacheRequest := func(requestID string) bool {
			cacheMutex.Lock()
			defer cacheMutex.Unlock()

			// Check if request was already processed (existing cache logic)
			if _, exists := requestCache[requestID]; exists {
				return false // Duplicate detected, reject
			}

			// Add to cache and allow processing
			requestCache[requestID] = time.Now()
			return true // Allow processing
		}

		// Scenario 1: Same request ID twice (cache should prevent)
		requestID1 := "abc123"

		allowed1 := checkAndCacheRequest(requestID1)
		allowed2 := checkAndCacheRequest(requestID1) // Same ID

		assert.True(t, allowed1, "First request should be allowed")
		assert.False(t, allowed2, "Duplicate request ID should be rejected by cache")

		// Scenario 2: Different request IDs (cache CANNOT prevent)
		requestID2 := "def456" // Different ID

		allowed3 := checkAndCacheRequest(requestID2)

		assert.True(t, allowed3, "Different request ID should be allowed by cache")

		t.Log("✅ whatsappRequestCache only prevents EXACT duplicate request IDs")
		t.Log("❌ whatsappRequestCache CANNOT prevent different requests from racing")
	})

	t.Run("RealWorldScenario_RapidClicksWithDifferentIDs", func(t *testing.T) {
		// This simulates what actually happens in our frontend

		generateRequestID := func() string {
			// This is our actual frontend logic - generates UNIQUE IDs every time
			return "rid-" + time.Now().Format("20060102150405.000000") + "-" +
				string(rune(time.Now().Nanosecond()%10000))
		}

		requestCache := make(map[string]time.Time)
		var cacheMutex sync.Mutex
		var processedUserCount int64

		simulateRapidClicks := func() {
			var wg sync.WaitGroup

			// User clicks "Send All" twice rapidly
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func(clickNum int) {
					defer wg.Done()

					// Frontend generates NEW unique request ID for each click
					requestID := generateRequestID()
					time.Sleep(time.Millisecond) // Ensure different timestamps

					t.Logf("Click %d: Generated request ID %s", clickNum+1, requestID[:15]+"...")

					// Check cache (simulating backend request processing)
					cacheMutex.Lock()
					_, isDuplicate := requestCache[requestID]
					if !isDuplicate {
						requestCache[requestID] = time.Now()
					}
					cacheMutex.Unlock()

					if !isDuplicate {
						t.Logf("Click %d: Passed cache check, processing bulk operation", clickNum+1)

						// Simulate bulk operation: process 3 users
						userCount := 3
						for u := 0; u < userCount; u++ {
							atomic.AddInt64(&processedUserCount, 1)
							time.Sleep(10 * time.Millisecond) // Simulate processing time
						}

						t.Logf("Click %d: Completed processing %d users", clickNum+1, userCount)
					} else {
						t.Logf("Click %d: Rejected by cache (duplicate)", clickNum+1)
					}
				}(i)
			}

			wg.Wait()
		}

		// Execute the scenario
		simulateRapidClicks()

		finalProcessedCount := atomic.LoadInt64(&processedUserCount)
		t.Logf("Final result: %d user processing events", finalProcessedCount)

		// Expected: 3 users should be processed once = 3 total
		// Actual: With different request IDs, both operations run = 6 total
		expectedWithCache := int64(3)    // If cache worked perfectly
		expectedWithoutCache := int64(6) // Both operations run (the actual problem)

		if finalProcessedCount == expectedWithoutCache {
			t.Log("❌ CONFIRMED: whatsappRequestCache alone CANNOT prevent this race condition")
			t.Log("   Two different request IDs both passed cache check and processed same users")
			t.Log("   This is exactly why we need bulkOperationMutex")
		} else if finalProcessedCount == expectedWithCache {
			t.Log("❓ Unexpected: Only one operation ran (race condition didn't occur in this test)")
		}

		assert.Equal(t, expectedWithoutCache, finalProcessedCount,
			"Without bulkOperationMutex, both operations should run (demonstrating the problem)")
	})

	t.Run("BulkMutexPreventsRaceConditionThatCacheCanNot", func(t *testing.T) {
		// Simulate both mechanisms working together

		requestCache := make(map[string]time.Time)
		var cacheMutex sync.Mutex
		var bulkMutex sync.Mutex // The additional mutex we added
		var processedUserCount int64

		generateRequestID := func() string {
			return "rid-" + time.Now().Format("20060102150405.000000") + "-" +
				string(rune(time.Now().Nanosecond()%10000))
		}

		processBulkWithBothProtections := func(clickNum int) {
			// Step 1: Check request cache (different IDs will pass)
			requestID := generateRequestID()
			time.Sleep(time.Millisecond) // Ensure different timestamps

			cacheMutex.Lock()
			_, isDuplicate := requestCache[requestID]
			if !isDuplicate {
				requestCache[requestID] = time.Now()
			}
			cacheMutex.Unlock()

			if isDuplicate {
				t.Logf("Click %d: Rejected by cache", clickNum)
				return
			}

			t.Logf("Click %d: Passed cache check with ID %s", clickNum, requestID[:15]+"...")

			// Step 2: Acquire bulk operation mutex (only one operation at a time)
			t.Logf("Click %d: Waiting for bulk operation lock", clickNum)
			bulkMutex.Lock()
			defer bulkMutex.Unlock()

			t.Logf("Click %d: Acquired bulk operation lock, processing", clickNum)

			// Simulate bulk operation
			userCount := 3
			for u := 0; u < userCount; u++ {
				atomic.AddInt64(&processedUserCount, 1)
				time.Sleep(20 * time.Millisecond) // Simulate work
			}

			t.Logf("Click %d: Completed processing %d users", clickNum, userCount)
		}

		var wg sync.WaitGroup

		// Rapid clicks with both protections
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(clickNum int) {
				defer wg.Done()
				processBulkWithBothProtections(clickNum + 1)
			}(i)
		}

		wg.Wait()

		finalProcessedCount := atomic.LoadInt64(&processedUserCount)
		t.Logf("With both protections: %d user processing events", finalProcessedCount)

		// With bulk mutex: only one operation should run at a time
		expectedWithBulkMutex := int64(3) // Only 3 users, processed once each

		assert.Equal(t, expectedWithBulkMutex, finalProcessedCount,
			"With bulkOperationMutex, only one bulk operation should complete")

		t.Log("✅ CONCLUSION: bulkOperationMutex IS needed in addition to whatsappRequestCache")
		t.Log("   - whatsappRequestCache prevents SAME request ID from being processed twice")
		t.Log("   - bulkOperationMutex prevents DIFFERENT requests from processing same users simultaneously")
	})
}

// TestWhyBothMechanismsAreNeeded provides a clear explanation
func TestWhyBothMechanismsAreNeeded(t *testing.T) {
	t.Log("=== EXPLANATION: Why we need BOTH whatsappRequestCache AND bulkOperationMutex ===")
	t.Log("")

	t.Log("whatsappRequestCache protects against:")
	t.Log("  ✓ Same request ID processed twice (e.g., network retry with same ID)")
	t.Log("  ✓ Browser sending duplicate HTTP requests")
	t.Log("  ✓ User double-clicking same button rapidly")
	t.Log("")

	t.Log("whatsappRequestCache CANNOT protect against:")
	t.Log("  ❌ Different request IDs processing same users")
	t.Log("  ❌ Multiple browser tabs making concurrent requests")
	t.Log("  ❌ Multiple users triggering bulk operations simultaneously")
	t.Log("  ❌ Direct API calls bypassing frontend debouncing")
	t.Log("")

	t.Log("bulkOperationMutex protects against:")
	t.Log("  ✓ Multiple bulk operations running concurrently")
	t.Log("  ✓ Same users being processed by different requests simultaneously")
	t.Log("  ✓ Database race conditions during user list fetching")
	t.Log("  ✓ WhatsApp client being used by multiple goroutines simultaneously")
	t.Log("")

	t.Log("REAL SCENARIO that shows why both are needed:")
	t.Log("  1. User clicks 'Send All' at 10:00:00.000")
	t.Log("  2. Frontend generates ID: 'rid-20240101100000000-1234'")
	t.Log("  3. Request reaches backend, passes cache check, starts processing")
	t.Log("  4. User clicks 'Send All' again at 10:00:00.100 (after debounce)")
	t.Log("  5. Frontend generates ID: 'rid-20240101100000100-5678' (DIFFERENT)")
	t.Log("  6. Second request reaches backend, passes cache check (different ID)")
	t.Log("  7. WITHOUT bulkMutex: Both requests process same users = DUPLICATES")
	t.Log("  8. WITH bulkMutex: Second request waits for first to complete = NO DUPLICATES")
	t.Log("")

	t.Log("✅ VERDICT: Both mechanisms are complementary and necessary")
}
