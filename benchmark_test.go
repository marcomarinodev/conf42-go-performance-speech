package main

import (
	"io"
	"net/http/httptest"
	"net/url"
	"testing"
)

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

func BenchmarkFiltering_Linear(b *testing.B) {
	// Generate sample transactions
	transactions := generateTransactionsForTest(100)

	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		_ = simpleFilterByPrefix(transactions, "USB")
	}
}

func BenchmarkFiltering_PrefixTree(b *testing.B) {
	// Generate sample transactions
	transactions := generateTransactionsForTest(100)

	b.ResetTimer()

	// Construct prefix tree
	trie := constructPrefixTree(transactions)

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		_ = filterByPrefixTree(trie, "USB")
	}
}
