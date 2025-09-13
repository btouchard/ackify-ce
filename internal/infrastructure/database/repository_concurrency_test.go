//go:build integration

package database

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

func TestRepository_Concurrency_Integration(t *testing.T) {
	testDB := SetupTestDB(t)
	repo := NewSignatureRepository(testDB.DB)
	factory := NewSignatureFactory()
	ctx := context.Background()

	t.Run("concurrent creates different docs", func(t *testing.T) {
		testDB.ClearTable(t)

		const numGoroutines = 50
		const signaturesPerGoroutine = 10

		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*signaturesPerGoroutine)

		// Launch concurrent goroutines creating signatures
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < signaturesPerGoroutine; j++ {
					sig := factory.CreateSignatureWithDocAndUser(
						fmt.Sprintf("doc-%d-%d", goroutineID, j),
						fmt.Sprintf("user-%d-%d", goroutineID, j),
						fmt.Sprintf("user%d%d@example.com", goroutineID, j),
					)

					if err := repo.Create(ctx, sig); err != nil {
						errors <- err
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent create error: %v", err)
		}

		// Verify all signatures were created
		expectedCount := numGoroutines * signaturesPerGoroutine
		actualCount := testDB.GetTableCount(t)
		if actualCount != expectedCount {
			t.Errorf("Expected %d signatures, got %d", expectedCount, actualCount)
		}
	})

	t.Run("concurrent creates with duplicate attempts", func(t *testing.T) {
		testDB.ClearTable(t)

		const numGoroutines = 20
		docID := "shared-doc"
		userSub := "shared-user"

		var wg sync.WaitGroup
		successCount := make(chan int, numGoroutines)
		errorCount := make(chan int, numGoroutines)

		// Launch concurrent goroutines trying to create the same signature
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				sig := factory.CreateSignatureWithDocAndUser(
					docID,
					userSub,
					"shared@example.com",
				)

				if err := repo.Create(ctx, sig); err != nil {
					errorCount <- 1
				} else {
					successCount <- 1
				}
			}()
		}

		wg.Wait()
		close(successCount)
		close(errorCount)

		// Count results
		successes := 0
		failures := 0
		for range successCount {
			successes++
		}
		for range errorCount {
			failures++
		}

		// Only one should succeed due to unique constraint
		if successes != 1 {
			t.Errorf("Expected exactly 1 success, got %d", successes)
		}
		if failures != numGoroutines-1 {
			t.Errorf("Expected %d failures, got %d", numGoroutines-1, failures)
		}

		// Verify only one record exists
		count := testDB.GetTableCount(t)
		if count != 1 {
			t.Errorf("Expected 1 signature after concurrent duplicates, got %d", count)
		}
	})

	t.Run("concurrent reads during writes", func(t *testing.T) {
		testDB.ClearTable(t)

		const numWriters = 10
		const numReaders = 20
		const numWrites = 5
		docID := "concurrent-doc"

		var wg sync.WaitGroup

		// Start writers
		for i := 0; i < numWriters; i++ {
			wg.Add(1)
			go func(writerID int) {
				defer wg.Done()

				for j := 0; j < numWrites; j++ {
					sig := factory.CreateSignatureWithDocAndUser(
						docID,
						fmt.Sprintf("user-%d-%d", writerID, j),
						fmt.Sprintf("user%d%d@example.com", writerID, j),
					)

					_ = repo.Create(ctx, sig)
					time.Sleep(time.Millisecond) // Small delay to spread writes
				}
			}(i)
		}

		// Start readers
		readResults := make(chan int, numReaders*10) // Buffer for all results
		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < 10; j++ {
					signatures, err := repo.GetByDoc(ctx, docID)
					if err != nil {
						t.Errorf("Concurrent read error: %v", err)
						return
					}
					readResults <- len(signatures)
					time.Sleep(time.Millisecond)
				}
			}()
		}

		wg.Wait()
		close(readResults)

		// Verify reads were consistent (no corruption)
		for count := range readResults {
			if count < 0 || count > numWriters*numWrites {
				t.Errorf("Invalid read result: %d (should be 0-%d)", count, numWriters*numWrites)
			}
		}

		// Verify final count
		finalCount := testDB.GetTableCount(t)
		expectedCount := numWriters * numWrites
		if finalCount != expectedCount {
			t.Errorf("Expected %d final signatures, got %d", expectedCount, finalCount)
		}
	})

	t.Run("concurrent GetLastSignature during creates", func(t *testing.T) {
		testDB.ClearTable(t)

		const numCreators = 10
		const numReaders = 5

		var wg sync.WaitGroup

		// Start creators
		for i := 0; i < numCreators; i++ {
			wg.Add(1)
			go func(creatorID int) {
				defer wg.Done()

				for j := 0; j < 5; j++ {
					sig := factory.CreateSignatureWithUser(
						fmt.Sprintf("user-%d-%d", creatorID, j),
						fmt.Sprintf("user%d%d@example.com", creatorID, j),
					)

					_ = repo.Create(ctx, sig)
					time.Sleep(2 * time.Millisecond)
				}
			}(i)
		}

		// Start readers calling GetLastSignature
		lastSigResults := make(chan *models.Signature, numReaders*10)
		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < 10; j++ {
					lastSig, err := repo.GetLastSignature(ctx)
					if err != nil {
						t.Errorf("GetLastSignature error: %v", err)
						return
					}
					lastSigResults <- lastSig
					time.Sleep(time.Millisecond)
				}
			}()
		}

		wg.Wait()
		close(lastSigResults)

		// Verify GetLastSignature results are valid
		for sig := range lastSigResults {
			if sig != nil {
				// Should have valid ID assigned by database
				if sig.ID <= 0 {
					t.Error("GetLastSignature returned signature with invalid ID")
				}
				// Should have valid required fields
				if sig.DocID == "" || sig.UserSub == "" {
					t.Error("GetLastSignature returned signature with empty required fields")
				}
			}
		}
	})

	t.Run("concurrent GetAllSignaturesOrdered during creates", func(t *testing.T) {
		testDB.ClearTable(t)

		const numCreators = 5
		const numReaders = 3

		var wg sync.WaitGroup

		// Start creators
		for i := 0; i < numCreators; i++ {
			wg.Add(1)
			go func(creatorID int) {
				defer wg.Done()

				for j := 0; j < 10; j++ {
					sig := factory.CreateSignatureWithUser(
						fmt.Sprintf("concurrent-user-%d-%d", creatorID, j),
						fmt.Sprintf("user%d%d@example.com", creatorID, j),
					)

					_ = repo.Create(ctx, sig)
					time.Sleep(time.Millisecond)
				}
			}(i)
		}

		// Start readers calling GetAllSignaturesOrdered
		orderingErrors := make(chan error, numReaders*5)
		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < 5; j++ {
					signatures, err := repo.GetAllSignaturesOrdered(ctx)
					if err != nil {
						orderingErrors <- err
						return
					}

					// Verify ordering (ID should be ascending)
					for k := 1; k < len(signatures); k++ {
						if signatures[k].ID <= signatures[k-1].ID {
							orderingErrors <- err
							return
						}
					}

					time.Sleep(5 * time.Millisecond)
				}
			}()
		}

		wg.Wait()
		close(orderingErrors)

		// Check for ordering violations
		for err := range orderingErrors {
			if err != nil {
				t.Errorf("Concurrent ordering error: %v", err)
			}
		}
	})

	t.Run("stress test with mixed operations", func(t *testing.T) {
		testDB.ClearTable(t)

		const duration = 2 * time.Second
		const numWorkers = 20

		ctx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()

		var wg sync.WaitGroup
		operationCounts := make(chan map[string]int, numWorkers)

		// Start workers doing mixed operations
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				counts := map[string]int{
					"creates": 0,
					"gets":    0,
					"exists":  0,
					"last":    0,
					"all":     0,
					"errors":  0,
				}

				for {
					select {
					case <-ctx.Done():
						operationCounts <- counts
						return
					default:
						// Randomly choose operation
						switch workerID % 5 {
						case 0: // Create
							sig := factory.CreateSignatureWithUser(
								fmt.Sprintf("stress-user-%d-%d", workerID, counts["creates"]),
								fmt.Sprintf("stress%d%d@example.com", workerID, counts["creates"]),
							)
							if err := repo.Create(ctx, sig); err != nil {
								counts["errors"]++
							} else {
								counts["creates"]++
							}

						case 1: // GetByDocAndUser
							_, err := repo.GetByDocAndUser(ctx, "test-doc-123", "user-123")
							if err != nil && !strings.Contains(err.Error(), "not found") {
								counts["errors"]++
							} else {
								counts["gets"]++
							}

						case 2: // ExistsByDocAndUser
							_, err := repo.ExistsByDocAndUser(ctx, "test-doc-123", "user-123")
							if err != nil {
								counts["errors"]++
							} else {
								counts["exists"]++
							}

						case 3: // GetLastSignature
							_, err := repo.GetLastSignature(ctx)
							if err != nil {
								counts["errors"]++
							} else {
								counts["last"]++
							}

						case 4: // GetAllSignaturesOrdered
							_, err := repo.GetAllSignaturesOrdered(ctx)
							if err != nil {
								counts["errors"]++
							} else {
								counts["all"]++
							}
						}
					}
				}
			}(i)
		}

		wg.Wait()
		close(operationCounts)

		// Aggregate results
		totalOps := 0
		totalErrors := 0
		for counts := range operationCounts {
			for op, count := range counts {
				if op == "errors" {
					totalErrors += count
				} else {
					totalOps += count
				}
			}
		}

		t.Logf("Stress test completed: %d operations, %d errors", totalOps, totalErrors)

		// Should have completed many operations with minimal errors
		if totalOps < 100 {
			t.Errorf("Expected at least 100 operations, got %d", totalOps)
		}

		// Error rate should be reasonable (< 10%)
		errorRate := float64(totalErrors) / float64(totalOps+totalErrors) * 100
		if errorRate > 10 {
			t.Errorf("Error rate too high: %.2f%% (expected < 10%%)", errorRate)
		}
	})
}

func TestRepository_DeadlockPrevention_Integration(t *testing.T) {
	testDB := SetupTestDB(t)
	factory := NewSignatureFactory()
	ctx := context.Background()

	t.Run("avoid deadlocks with multiple table access patterns", func(t *testing.T) {
		testDB.ClearTable(t)

		const numWorkers = 10
		const opsPerWorker = 20

		var wg sync.WaitGroup
		deadlockErrors := make(chan error, numWorkers)

		// Workers with different access patterns that could cause deadlocks
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				repo := NewSignatureRepository(testDB.DB)

				for j := 0; j < opsPerWorker; j++ {
					// Pattern 1: Create then immediately query
					if workerID%2 == 0 {
						sig := factory.CreateSignatureWithUser(
							fmt.Sprintf("pattern1-user-%d-%d", workerID, j),
							fmt.Sprintf("pattern1-%d%d@example.com", workerID, j),
						)

						if err := repo.Create(ctx, sig); err == nil {
							_, _ = repo.GetByDocAndUser(ctx, sig.DocID, sig.UserSub)
							_, _ = repo.ExistsByDocAndUser(ctx, sig.DocID, sig.UserSub)
						}
					} else {
						// Pattern 2: Query then create
						testDocID := fmt.Sprintf("pattern2-doc-%d", workerID)
						testUserSub := fmt.Sprintf("pattern2-user-%d", j)

						_, _ = repo.GetByDoc(ctx, testDocID)
						_, _ = repo.GetByUser(ctx, testUserSub)

						sig := factory.CreateSignatureWithDocAndUser(
							testDocID,
							testUserSub,
							"pattern2@example.com",
						)
						_ = repo.Create(ctx, sig)
					}

					// Small random delay to increase chance of contention
					time.Sleep(time.Duration(workerID%3+1) * time.Millisecond)
				}
			}(i)
		}

		// Wait with timeout to detect deadlocks
		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
			// Success - no deadlocks
		case <-time.After(30 * time.Second):
			t.Fatal("Test timed out - possible deadlock detected")
		}

		close(deadlockErrors)

		// Check for deadlock-specific errors
		for err := range deadlockErrors {
			if err != nil {
				t.Errorf("Deadlock-related error: %v", err)
			}
		}
	})
}
