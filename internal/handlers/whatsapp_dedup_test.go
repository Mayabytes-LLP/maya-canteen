package handlers

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestClaimNotificationKey(t *testing.T) {
	const key = "test-claim-key"
	releaseNotificationKeys(key)
	t.Cleanup(func() { releaseNotificationKeys(key) })

	if !claimNotificationKey(key) {
		t.Fatal("first claim should succeed")
	}
	if claimNotificationKey(key) {
		t.Fatal("second claim of the same key must be reported as duplicate")
	}
	releaseNotificationKeys(key, "")
	if !claimNotificationKey(key) {
		t.Fatal("claim after release should succeed (retry allowed)")
	}
}

func TestClaimNotificationKeyConcurrent(t *testing.T) {
	const key = "test-concurrent-key"
	releaseNotificationKeys(key)
	t.Cleanup(func() { releaseNotificationKeys(key) })

	var winners atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if claimNotificationKey(key) {
				winners.Add(1)
			}
		}()
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("exactly one concurrent claim must win, got %d", winners.Load())
	}
}

func TestBulkMutexRejectsSecondRun(t *testing.T) {
	if !bulkOperationMutex.TryLock() {
		t.Fatal("bulk lock should be free")
	}
	if bulkOperationMutex.TryLock() {
		t.Fatal("second bulk run must not acquire the lock while the first holds it")
	}
	bulkOperationMutex.Unlock()
}
