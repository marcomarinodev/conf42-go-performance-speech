package tests

import (
	"io"
	"net/http/httptest"
	"net/url"
	"testing"
	"transactions_sample/pkg/transaction"
)

func BenchmarkRequestSliceHandler_WithoutCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "false")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		transaction.ProcessTransactionPipeline(rr, req)

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
		transaction.ProcessTransactionPipeline(rr, req)

		if i == 0 {
			_, err := io.ReadAll(rr.Body)
			if err != nil {
				b.Fatalf("Error reading response body: %v", err)
			}

		}
	}
}
