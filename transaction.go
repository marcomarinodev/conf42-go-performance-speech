package main

import (
	"strings"
	"sync"
	"time"
)

type Transaction struct {
	TransactionID string    `bson:"transactionID"`
	Timestamp     time.Time `bson:"timestamp"`
	CustomerID    string    `bson:"customerID"`
	ProductName   string    `bson:"productName"`
	Category      string    `bson:"category"`
	Quantity      int       `bson:"quantity"`
	UnitPrice     float64   `bson:"unitPrice"`
	TotalAmount   float64   `bson:"totalAmount"`
	PaymentMethod string    `bson:"paymentMethod"`
}

type AggregatedTransaction struct {
	Category      string
	TotalQuantity int
	TotalAmount   float64
	Count         int
}

type ProcessedTransaction struct {
	Category    string
	TotalSales  float64
	AvgQuantity float64
}

// === slice and map simple filtering ===
func simpleFilterByPrefixFromSlice(transactions []Transaction, prefix string) []Transaction {
	filteredTransactions := make([]Transaction, 0)
	for _, transaction := range transactions {
		if strings.HasPrefix(transaction.ProductName, prefix) {
			filteredTransactions = append(filteredTransactions, transaction)
		}
	}
	return filteredTransactions
}

func testFilterChan(transactions []Transaction, prefix string) []Transaction {
	filteredTransactions := make([]Transaction, 0)
	for _, transaction := range transactions {
		if strings.HasPrefix(transaction.ProductName, prefix) {
			filteredTransactions = append(filteredTransactions, transaction)
		}
	}

	return filteredTransactions
}

func parallelFilterByPrefixFromSlice(transactions []Transaction, prefix string, n int) []Transaction {
	var wg sync.WaitGroup

	res := make(chan []Transaction)

	// Calculate the number of transactions per goroutine
	transactionsPerRoutine := len(transactions) / n

	// Launch goroutines
	for i := 0; i < n; i++ {
		wg.Add(1)
		startIndex := i * transactionsPerRoutine
		endIndex := (i + 1) * transactionsPerRoutine
		if i == n-1 {
			// Ensure the last goroutine handles any remaining transactions
			endIndex = len(transactions)
		}
		go func(start, end int) {
			defer wg.Done()
			res <- testFilterChan(transactions[start:end], prefix)
		}(startIndex, endIndex)
	}

	// Goroutine to close the result channel once all goroutines are done
	go func() {
		wg.Wait()
		close(res)
	}()

	var result []Transaction
	for r := range res {
		result = append(result, r...)
	}
	return result
}

func simpleFilterByPrefixFromMap(transactions map[string]Transaction, prefix string) []Transaction {
	filteredTransactions := make([]Transaction, 0)
	for _, transaction := range transactions {
		if strings.HasPrefix(transaction.ProductName, prefix) {
			filteredTransactions = append(filteredTransactions, transaction)
		}
	}
	return filteredTransactions
}

// === sequential and parallel aggregate ===
func aggregateTransactions(transactions []Transaction) []AggregatedTransaction {
	aggregatedTransactions := make(map[string]AggregatedTransaction)

	// Aggregate transactions
	for _, transaction := range transactions {
		// Retrieve the aggregated transaction from the map
		aggTransaction, ok := aggregatedTransactions[transaction.Category]
		if !ok {
			// If the category doesn't exist, create a new aggregated transaction
			aggTransaction = AggregatedTransaction{
				Category:      transaction.Category,
				TotalQuantity: 0,
				TotalAmount:   0,
				Count:         0,
			}
		}

		// Update the aggregated transaction with the current transaction data
		aggTransaction.Category = transaction.Category
		aggTransaction.TotalQuantity += transaction.Quantity
		aggTransaction.TotalAmount += transaction.TotalAmount
		aggTransaction.Count++

		// Store the aggregated transaction back into the map
		aggregatedTransactions[transaction.Category] = aggTransaction
	}

	// Convert map to slice
	var result []AggregatedTransaction
	for _, aggTransaction := range aggregatedTransactions {
		result = append(result, aggTransaction)
	}

	return result
}

func parallelAggregateTransactions(input <-chan []Transaction, results chan<- []AggregatedTransaction) {

	filteredTransactions := make([]Transaction, 0)

	for filteredTransaction := range input {
		filteredTransactions = append(filteredTransactions, filteredTransaction...)
	}

	results <- aggregateTransactions(filteredTransactions)
}

func processTransactions(aggregatedTransactions []AggregatedTransaction) []ProcessedTransaction {
	processedTransactions := make([]ProcessedTransaction, 0)

	for _, aggregatedTransaction := range aggregatedTransactions {
		processedTransaction := ProcessedTransaction{
			Category:    "Processed_" + aggregatedTransaction.Category,
			TotalSales:  aggregatedTransaction.TotalAmount,
			AvgQuantity: float64(aggregatedTransaction.TotalQuantity) / float64(aggregatedTransaction.Count),
		}

		processedTransactions = append(processedTransactions, processedTransaction)
	}

	return processedTransactions
}
