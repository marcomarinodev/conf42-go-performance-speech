package main

import (
	"io"
	"net/http/httptest"
	"net/url"
	"testing"
)

func BenchmarkRequestHandlerWithoutCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "false")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		requestHandler(rr, req)

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

func BenchmarkRequestHandlerWithCache(b *testing.B) {
	query := url.Values{}
	query.Set("customerID", "CUST1")
	query.Set("withCache", "true")

	req := httptest.NewRequest("GET", "/transactions?"+query.Encode(), nil)
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		requestHandler(rr, req)

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

// func BenchmarkSlices(b *testing.B) {
// 	ctx := context.Background()

// }
