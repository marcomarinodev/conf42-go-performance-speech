package main

import (
	"testing"
)

func BenchmarkFetchCustomerTransactionsWithoutCache(b *testing.B) {
	customerId := "customer1"

	// Benchmark
	for i := 0; i < b.N; i++ {
		_, err := getTransactionsByCustomerID(customerId, false)
		if err != nil {
			b.Fatalf("failed to get transactions: %v", err)
		}
	}
}

func BenchmarkFetchCustomerTransactionsWithCache(b *testing.B) {
	customerId := "customer1"

	// Benchmark
	for i := 0; i < b.N; i++ {
		_, err := getTransactionsByCustomerID(customerId, true)
		if err != nil {
			b.Fatalf("failed to get transactions: %v", err)
		}
	}
}
