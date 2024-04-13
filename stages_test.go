package main

import (
	"io"
	"math"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

// ==== Benchmarking ====
func BenchmarkPipeline(b *testing.B) {
	allTransactions := generateTransactionsForTest(10000 * 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filteredTransactions := FilterByPrefix_seq(allTransactions, "US")

		res := AggregateTransactions_seq(filteredTransactions)

		_ = ProcessTransactions(res)
	}

}

// ? remember to run this benchmark with the same name as the non optimized version
func BenchmarkPipeline_Improved(b *testing.B) {
	allTransactions := generateTransactionsForTest(10000 * 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filteredTransactions := FilterByPrefix_par(allTransactions, "US", 4)

		res := AggregateTransactions_par(filteredTransactions, 4)

		_ = ProcessTransactions(res)
	}

}

func BenchmarkRequestSliceHandler_WithoutCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "false")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		processTransactionPipeline(rr, req)

		if i == 0 {
			_, err := io.ReadAll(rr.Body)
			if err != nil {
				b.Fatalf("Error reading response body: %v", err)
			}
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
		processTransactionPipeline(rr, req)

		if i == 0 {
			_, err := io.ReadAll(rr.Body)
			if err != nil {
				b.Fatalf("Error reading response body: %v", err)
			}

		}
	}
}

func BenchmarkFiltering(b *testing.B) {

	runBenchmark := func(b *testing.B, transactions []Transaction) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = FilterByPrefix_seq(transactions, "USB")
		}
	}

	runParallelBenchmark := func(b *testing.B, transactions []Transaction) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = FilterByPrefix_par(transactions, "USB", 4)
		}
	}

	b.Run("Linear5k", func(b *testing.B) {
		transactions := generateTransactionsForTest(5 * 1000)
		runBenchmark(b, transactions)
	})

	b.Run("Parallel5k", func(b *testing.B) {
		transactions := generateTransactionsForTest(5 * 1000)
		runParallelBenchmark(b, transactions)
	})

	b.Run("Linear100k", func(b *testing.B) {
		transactions := generateTransactionsForTest(1000 * 100)
		runBenchmark(b, transactions)
	})

	b.Run("Parallel100k", func(b *testing.B) {
		transactions := generateTransactionsForTest(1000 * 100)
		runParallelBenchmark(b, transactions)
	})
}

func BenchmarkAggregation(b *testing.B) {
	runBenchmark := func(b *testing.B, transactions []Transaction) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = AggregateTransactions_seq(transactions)
		}
	}

	runParallelBenchmark := func(b *testing.B, transactions []Transaction) {
		transactionsChan := formTransactionsChunksChannel(transactions, 4)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = AggregateTransactions_par(transactionsChan, 4)
		}
	}

	// Benchmark linear filtering for 5k transactions
	b.Run("Linear5k", func(b *testing.B) {
		transactions := generateTransactionsForTest(5 * 1000)
		runBenchmark(b, transactions)
	})

	b.Run("Parallel5k", func(b *testing.B) {
		transactions := generateTransactionsForTest(5 * 1000)
		runParallelBenchmark(b, transactions)
	})

	b.Run("Linear100k", func(b *testing.B) {
		transactions := generateTransactionsForTest(1000 * 100)
		runBenchmark(b, transactions)
	})

	b.Run("Parallel100k", func(b *testing.B) {
		transactions := generateTransactionsForTest(1000 * 100)
		runParallelBenchmark(b, transactions)
	})
}

// ===== Tests =====

// *** filtering tests ***
func TestFiltering_Parallel(t *testing.T) {
	transactions := generateTransactionsForTest(1 * 100)

	expected := FilterByPrefix_seq(transactions, "USB")

	for i := 1; i <= 10; i++ {
		respChan := FilterByPrefix_par(transactions, "USB", 4)

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

	expected := AggregateTransactions_seq(transactions)

	for i := 1; i <= 10; i++ {
		transactionsChan := formTransactionsChunksChannel(transactions, 10)
		res := AggregateTransactions_par(transactionsChan, 4)

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
