package main

import (
	"context"
	"io"
	"math/rand"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func BenchmarkTransacionsRetrieval(b *testing.B) {
	b.Run("SliceHandler", func(b *testing.B) {
		b.Run("WithoutCache", BenchmarkRequestSliceHandlerWithoutCache)
		b.Run("WithCache", BenchmarkRequestSliceHandlerWithCache)
	})

	b.Run("MapHandler", func(b *testing.B) {
		b.Run("WithoutCache", BenchmarkRequestMapHandlerWithoutCache)
		b.Run("WithCache", BenchmarkRequestMapHandlerWithCache)
	})
}

func BenchmarkTransactionsProcessor(b *testing.B) {
	b.Run("SliceProcessor", func(b *testing.B) {
		b.Run("CalculateRevenue", BenchmarkCalculateTotalRevenueSlice)
		b.Run("Insert", BenchmarkInsertionSlice)
		b.Run("Seeking", BenchmarkSeekSlice)
	})

	b.Run("MapProcessor", func(b *testing.B) {
		b.Run("CalculateRevenue", BenchmarkCalculateTotalRevenueMap)
		b.Run("Insert", BenchmarkInsertionMap)
		b.Run("Seeking", BenchmarkSeekMap)
	})
}

func BenchmarkRequestSliceHandlerWithoutCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "false")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		requestSliceHandler(rr, req)

		if i == 0 {
			// Print the response body
			respBody, err := io.ReadAll(rr.Body)
			if err != nil {
				b.Fatalf("Error reading response body: %v", err)
			}

			b.Logf("Body n. of bytes: %d\n", len(respBody))
		}
	}
}

func BenchmarkRequestSliceHandlerWithCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "true")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		requestSliceHandler(rr, req)

		if i == 0 {
			// Print the response body
			respBody, err := io.ReadAll(rr.Body)
			if err != nil {
				b.Fatalf("Error reading response body: %v", err)
			}

			b.Logf("Body n. of bytes: %d\n", len(respBody))
		}
	}
}

func BenchmarkRequestMapHandlerWithoutCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "false")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		requestMapHandler(rr, req)

		if i == 0 {
			// Print the response body
			respBody, err := io.ReadAll(rr.Body)
			if err != nil {
				b.Fatalf("Error reading response body: %v", err)
			}

			b.Logf("Body n. of bytes: %d\n", len(respBody))
		}
	}
}

func BenchmarkRequestMapHandlerWithCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "true")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		requestMapHandler(rr, req)

		if i == 0 {
			// Print the response body
			respBody, err := io.ReadAll(rr.Body)
			if err != nil {
				b.Fatalf("Error reading response body: %v", err)
			}

			b.Logf("Body n. of bytes: %d\n", len(respBody))
		}
	}
}

func BenchmarkCalculateTotalRevenueSlice(b *testing.B) {
	transactions, err := getSliceFromDb(context.Background(), "CUST1")

	if err != nil {
		b.Fatalf("Error retrieving transactions: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateTotalRevenueSlice(transactions)
	}
}

func BenchmarkCalculateTotalRevenueMap(b *testing.B) {
	transactions, err := getMapFromDb(context.Background(), "CUST1")

	if err != nil {
		b.Fatalf("Error retrieving transactions: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateTotalRevenueMap(transactions)
	}
}

func generateRandomTransaction() Transaction {
	return Transaction{
		TransactionID: "id" + randString(10),
		Timestamp:     time.Now(),
		CustomerID:    "customer" + randString(5),
		ProductName:   "product" + randString(5),
		Category:      "category" + randString(5),
		Quantity:      rand.Intn(100),
		UnitPrice:     rand.Float64() * 100,
		TotalAmount:   rand.Float64() * 1000,
		PaymentMethod: "payment" + randString(5),
	}
}

func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func BenchmarkInsertionMap(b *testing.B) {
	transactions := make(map[string]Transaction)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := generateRandomTransaction()
		transactions[tx.TransactionID] = tx
	}
}

func BenchmarkInsertionSlice(b *testing.B) {
	var transactions []Transaction
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := generateRandomTransaction()
		transactions = append(transactions, tx)
	}
}

func BenchmarkSeekMap(b *testing.B) {
	transactions := make(map[string]Transaction)
	for i := 0; i < 10000; i++ {
		tx := generateRandomTransaction()
		transactions[tx.TransactionID] = tx
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = transactions["id"+randString(10)] // Seeking a random transaction
	}
}

func BenchmarkSeekSlice(b *testing.B) {
	var transactions []Transaction
	for i := 0; i < 10000; i++ {
		tx := generateRandomTransaction()
		transactions = append(transactions, tx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findTransactionByID(transactions, "id"+randString(10)) // Seeking a random transaction
	}
}
