package main

import (
	"io"
	"math"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"
)

// ==== Benchmarking ====
func BenchmarkRequestSliceHandler_WithoutCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "false")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		processTransactionsSlice(rr, req)

		if i == 0 {
			// Print the response body
			_, err := io.ReadAll(rr.Body)
			if err != nil {
				b.Fatalf("Error reading response body: %v", err)
			}

			// b.Logf("Body n. of bytes: %d\n", len(respBody))
		}
	}
}

func BenchmarkRequestSliceHandler_WithCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "true")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		processTransactionsSlice(rr, req)

		if i == 0 {
			// Print the response body
			_, err := io.ReadAll(rr.Body)
			if err != nil {
				b.Fatalf("Error reading response body: %v", err)
			}

			// b.Logf("Body n. of bytes: %d\n", len(respBody))
		}
	}
}

func BenchmarkFiltering(b *testing.B) {
	// Define benchmarking logic for linear and parallel filtering
	runBenchmark := func(b *testing.B, transactions []Transaction) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = filterByPrefix_seq(transactions, "USB")
		}
	}

	runParallelBenchmark := func(b *testing.B, transactions []Transaction) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = filterByPrefix_par(transactions, "USB", 4)
		}
	}

	// Benchmark linear filtering for 5k transactions
	b.Run("Linear5k", func(b *testing.B) {
		transactions := generateTransactionsForTest(5 * 1000)
		runBenchmark(b, transactions)
	})

	// Benchmark parallel filtering for 5k transactions
	b.Run("Parallel5k", func(b *testing.B) {
		transactions := generateTransactionsForTest(5 * 1000)
		runParallelBenchmark(b, transactions)
	})

	// Benchmark linear filtering for 100k transactions
	b.Run("Linear100k", func(b *testing.B) {
		transactions := generateTransactionsForTest(1000 * 100)
		runBenchmark(b, transactions)
	})

	// Benchmark parallel filtering for 100k transactions
	b.Run("Parallel100k", func(b *testing.B) {
		transactions := generateTransactionsForTest(1000 * 100)
		runParallelBenchmark(b, transactions)
	})
}

// ===== Tests =====

// *** filtering tests ***
func TestFiltering_Parallel(t *testing.T) {
	transactions := generateTransactionsForTest(1 * 100)

	expected := filterByPrefix_seq(transactions, "USB")

	for i := 1; i <= 10; i++ {
		respChan := filterByPrefix_par(transactions, "USB", 4)

		var actual []Transaction
		for res := range respChan {
			actual = append(actual, res...)
		}

		if !EqualTransactions(expected, actual) {
			t.Errorf("Transactions slices are not equal")
		}
	}
}

// *** aggregation tests ***
func TestAggregation_Parallel(t *testing.T) {
	transactions := generateTransactionsForTest(1 * 100)

	expected := aggregateTransactions_seq(transactions)

	for i := 1; i <= 10; i++ {
		transactionsChan := formTransactionsChunksChannel(transactions, 10)
		aggResultChan := make(chan []AggregatedTransaction)
		aggNumWorkers := 4

		// fan-out: create worker goroutines
		for i := 0; i < aggNumWorkers; i++ {
			go aggregatorWorker(i, transactionsChan, aggResultChan)
		}

		// fan-in: collect results
		var aggWg sync.WaitGroup
		aggWg.Add(aggNumWorkers)

		go func() {
			aggWg.Wait()         // Wait for all jobs to be done
			close(aggResultChan) // Close the results channel after all jobs are processed
		}()

		res := mergeAggregatedTransactions(aggResultChan, &aggWg)

		if !EqualAggregatedTransactions(expected, res) {
			t.Errorf("Transactions slices are not equal")
		}
	}
}

// ===== helper functions =====
func EqualTransactions(slice1, slice2 []Transaction) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	// Create maps to count occurrences of each transaction in both slices
	count1 := make(map[Transaction]int)
	count2 := make(map[Transaction]int)

	for _, t := range slice1 {
		count1[t]++
	}
	for _, t := range slice2 {
		count2[t]++
	}

	// Compare the counts of transactions in both maps
	return reflect.DeepEqual(count1, count2)
}

func EqualAggregatedTransactions(sliceT []AggregatedTransaction, mapT map[string]AggregatedTransaction) bool {
	if len(sliceT) != len(mapT) {
		return false
	}

	// Create maps to count occurrences of each transaction in both slices
	count1 := make(map[AggregatedTransaction]int)
	count2 := make(map[AggregatedTransaction]int)

	for _, t := range sliceT {
		roundedT := t
		roundedT.TotalAmount = reduceToTwoDigits(roundedT.TotalAmount)
		count1[roundedT]++
	}
	for _, t := range mapT {
		roundedT := t
		roundedT.TotalAmount = reduceToTwoDigits(roundedT.TotalAmount)
		count2[roundedT]++
	}

	// Compare the counts of transactions in both maps
	return reflect.DeepEqual(count1, count2)
}

func formTransactionsChunksChannel(transactions []Transaction, chunkSize int) chan []Transaction {
	ch := make(chan []Transaction)

	go func() {
		defer close(ch)

		for i := 0; i < len(transactions); i += chunkSize {
			end := i + chunkSize
			if end > len(transactions) {
				end = len(transactions)
			}
			ch <- transactions[i:end]
		}
	}()

	return ch
}

func reduceToTwoDigits(num float64) float64 {
	return math.Round(num*100) / 100
}
