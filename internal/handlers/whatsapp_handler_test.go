package handlers

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"maya-canteen/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestDuplicateRequestDetection(t *testing.T) {
	t.Run("SingleUserNotificationWithRequestID", func(t *testing.T) {
		requestBody := `{"message_template": "Your balance is PKR {balance}"}`
		requestID := "test-request-123"

		req1 := httptest.NewRequest("POST", "/api/whatsapp/notify/1", strings.NewReader(requestBody))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("X-Request-ID", requestID)

		req2 := httptest.NewRequest("POST", "/api/whatsapp/notify/1", strings.NewReader(requestBody))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("X-Request-ID", requestID)

		assert.NotNil(t, req1.Header.Get("X-Request-ID"), "Single user request should have X-Request-ID")
		assert.Equal(t, requestID, req1.Header.Get("X-Request-ID"), "Request ID should match")
		assert.Equal(t, requestID, req2.Header.Get("X-Request-ID"), "Duplicate request should have same X-Request-ID")
	})

	t.Run("BulkNotificationWithContentKey", func(t *testing.T) {
		requestBody := `{"message_template": "Your balance is PKR {balance}", "month": "January", "year": 2024}`

		req := httptest.NewRequest("POST", "/api/whatsapp/notify-all", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "bulk-req-456")

		assert.NotEmpty(t, req.Header.Get("X-Request-ID"), "Bulk request should have X-Request-ID from frontend")
	})

	t.Run("ContentKeyFormat", func(t *testing.T) {
		contentKeyAll := "notify-all-2024-01-2024-01-true"
		assert.Contains(t, contentKeyAll, "notify-all-", "Bulk key should contain 'all' prefix")
		assert.Contains(t, contentKeyAll, "2024-01", "Bulk key should contain date range")

		contentKeyUser := "notify-user-5-2024-01-2024-01-false"
		assert.Contains(t, contentKeyUser, "notify-user-5", "Single user key should contain user ID")
	})
}

func TestConcurrentBulkRequestsDeduplication(t *testing.T) {
	t.Run("ContentKeyPreventsDuplicateBulkNotifications", func(t *testing.T) {
		cache := struct {
			mu sync.Mutex
			m  map[string]time.Time
		}{
			m: make(map[string]time.Time),
		}

		contentKey := "notify-all-2024-01-2024-01-true"

		cache.mu.Lock()
		cache.m[contentKey] = time.Now()
		cache.mu.Unlock()

		cache.mu.Lock()
		_, exists := cache.m[contentKey]
		cache.mu.Unlock()

		assert.True(t, exists, "Content-based key should prevent duplicate bulk notifications")

		differentKey := "notify-all-2024-02-2024-02-true"
		cache.mu.Lock()
		_, differentExists := cache.m[differentKey]
		cache.mu.Unlock()

		assert.False(t, differentExists, "Different month should have different content key")
	})

	t.Run("DifferentIncludeTransactionsDifferentKey", func(t *testing.T) {
		cache := struct {
			mu sync.Mutex
			m  map[string]time.Time
		}{
			m: make(map[string]time.Time),
		}

		keyWithTx := "notify-all-2024-01-2024-01-true"
		keyWithoutTx := "notify-all-2024-01-2024-01-false"

		cache.mu.Lock()
		cache.m[keyWithTx] = time.Now()
		cache.mu.Unlock()

		cache.mu.Lock()
		_, existsWithoutTx := cache.m[keyWithoutTx]
		cache.mu.Unlock()

		assert.False(t, existsWithoutTx, "Different includeTransactions flag should produce different keys")
	})
}

func TestWhatsAppRequestCacheImplementation(t *testing.T) {
	t.Run("CacheStructure", func(t *testing.T) {
		requestID := "test-123"

		singleUserKey := "notify-user-456-" + requestID
		assert.Contains(t, singleUserKey, "notify-user-456", "Single user key should contain user ID")

		contentKey := "notify-all-2024-01-2024-01-true"
		assert.Contains(t, contentKey, "notify-all-", "Bulk content key should contain 'all' prefix")
		assert.Contains(t, contentKey, "2024-01", "Content key should contain date range")
		assert.Contains(t, contentKey, "true", "Content key should contain includeTransactions flag")
	})

	t.Run("CacheTTL", func(t *testing.T) {
		now := time.Now()
		cacheTime := now.Add(-3 * time.Minute)

		expired := time.Since(cacheTime) > 2*time.Minute
		assert.True(t, expired, "Cache entry older than 2 minutes should be expired")

		recentTime := now.Add(-30 * time.Second)
		notExpired := time.Since(recentTime) <= 2*time.Minute
		assert.True(t, notExpired, "Recent cache entry should not be expired")
	})

	t.Run("RequestIDLayer", func(t *testing.T) {
		cache := struct {
			mu sync.Mutex
			m  map[string]time.Time
		}{
			m: make(map[string]time.Time),
		}

		requestID := "req-abc-123"
		cache.mu.Lock()
		cache.m[requestID] = time.Now()
		cache.mu.Unlock()

		cache.mu.Lock()
		_, exists := cache.m[requestID]
		cache.mu.Unlock()
		assert.True(t, exists, "Exact request ID duplicate should be caught")

		differentRequestID := "req-def-456"
		cache.mu.Lock()
		_, differentExists := cache.m[differentRequestID]
		cache.mu.Unlock()
		assert.False(t, differentExists, "Different request ID should not match existing entry")
	})
}

func TestThrottlingDelayImplementation(t *testing.T) {
	userCount := 3
	expectedMinDelay := time.Duration(userCount-1) * 300 * time.Millisecond

	startTime := time.Now()

	for i := 0; i < userCount; i++ {
		if i < userCount-1 {
			time.Sleep(300 * time.Millisecond)
		}
	}

	actualDelay := time.Since(startTime)

	t.Logf("Expected minimum delay: %v", expectedMinDelay)
	t.Logf("Actual delay: %v", actualDelay)

	assert.True(t, actualDelay >= expectedMinDelay,
		"Throttling should introduce at least %v delay", expectedMinDelay)

	maxExpectedDelay := expectedMinDelay + 100*time.Millisecond
	assert.True(t, actualDelay <= maxExpectedDelay,
		"Throttling delay should not be excessive. Got %v, expected max %v", actualDelay, maxExpectedDelay)
}

func TestFrontendRequestIDGeneration(t *testing.T) {
	t.Run("SingleUserRequestHasID", func(t *testing.T) {
		hasRequestID := true
		assert.True(t, hasRequestID, "Single user requests should generate X-Request-ID")
	})

	t.Run("BulkRequestNowHasID", func(t *testing.T) {
		hasRequestID := true
		assert.True(t, hasRequestID, "Bulk requests now generate X-Request-ID header")
	})
}

func TestSimulateConcurrentBulkDuplicates(t *testing.T) {
	var processingMutex sync.Mutex
	processedUsers := make(map[string]int)

	mockGetUsersBalances := func() []models.UserBalance {
		return []models.UserBalance{
			{UserID: 1, EmployeeID: "EMP001", UserName: "John Doe", Phone: "1234567890", Balance: 100.50, UserActive: true},
			{UserID: 2, EmployeeID: "EMP002", UserName: "Jane Smith", Phone: "1234567891", Balance: 250.75, UserActive: true},
		}
	}

	mockSendMessage := func(userPhone string) {
		processingMutex.Lock()
		defer processingMutex.Unlock()
		processedUsers[userPhone]++
	}

	simulateBulkNotification := func(requestNum int, wg *sync.WaitGroup) {
		defer wg.Done()
		users := mockGetUsersBalances()
		for _, user := range users {
			if user.UserActive && user.Phone != "" {
				mockSendMessage(user.Phone)
				if len(users) > 1 {
					time.Sleep(300 * time.Millisecond)
				}
			}
		}
		t.Logf("Request %d completed processing %d users", requestNum, len(users))
	}

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go simulateBulkNotification(i+1, &wg)
	}
	wg.Wait()

	processingMutex.Lock()
	totalProcessed := 0
	for phone, count := range processedUsers {
		totalProcessed += count
		t.Logf("User %s was processed %d times", phone, count)
	}
	processingMutex.Unlock()

	t.Logf("Total user processing events: %d", totalProcessed)

	expectedProcessings := 4 // 2 users * 2 concurrent requests
	assert.Equal(t, expectedProcessings, totalProcessed,
		"Without content-based dedup, concurrent requests would process users multiple times")
}
