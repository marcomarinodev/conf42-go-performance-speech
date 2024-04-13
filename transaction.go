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

// === filtering ===
func FilterByPrefix_seq(transactions []Transaction, prefix string) []Transaction {
	filteredTransactions := make([]Transaction, 0)
	for _, transaction := range transactions {
		if strings.HasPrefix(transaction.ProductName, prefix) {
			filteredTransactions = append(filteredTransactions, transaction)
		}
	}
	return filteredTransactions
}

func FilterByPrefix_par(transactions []Transaction, prefix string, numWorkers int) chan []Transaction {
	respChan := make(chan []Transaction, numWorkers)
	wg := &sync.WaitGroup{}
	partSize := len(transactions) / numWorkers
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go filterRoutine(transactions[i*partSize:(i+1)*partSize], prefix, respChan, wg)
	}

	wg.Wait()
	close(respChan)
	return respChan
}

func filterRoutine(transactions []Transaction, prefix string, respChan chan []Transaction, wg *sync.WaitGroup) {
	respChan <- FilterByPrefix_seq(transactions, prefix)
	wg.Done()
}

// === sequential and parallel aggregate ===
func AggregateTransactions_seq(transactions []Transaction) []AggregatedTransaction {
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

func AggregateTransactions_par(filteredTransactionsChan chan []Transaction, numWorkers int) map[string]AggregatedTransaction {
	aggResultChan := make(chan []AggregatedTransaction)

	// fan-out: create worker goroutines
	for i := 0; i < numWorkers; i++ {
		go aggregatorWorker(i, filteredTransactionsChan, aggResultChan)
	}

	// fan-in: collect results
	var aggWg sync.WaitGroup
	aggWg.Add(numWorkers)

	go func() {
		aggWg.Wait()         // Wait for all jobs to be done
		close(aggResultChan) // Close the results channel after all jobs are processed
	}()

	return mergeAggregatedTransactions(aggResultChan, &aggWg)
}

func aggregatorWorker(id int, transactions <-chan []Transaction, results chan<- []AggregatedTransaction) {
	aggregatedTransactions := make([]AggregatedTransaction, 0)

	for transaction := range transactions {
		aggregatedTransactions = append(aggregatedTransactions, AggregateTransactions_seq(transaction)...)
	}

	results <- aggregatedTransactions
}

func mergeAggregatedTransactions(aggTransactionsChan chan []AggregatedTransaction, aggregatorWg *sync.WaitGroup) map[string]AggregatedTransaction {
	mergedAggregatedTransactions := make(map[string]AggregatedTransaction)

	for aggTransactions := range aggTransactionsChan {
		for _, aggTransaction := range aggTransactions {
			// Retrieve the aggregated transaction from the map
			existingAggregatedTransaction, ok := mergedAggregatedTransactions[aggTransaction.Category]
			if !ok {
				// If the category doesn't exist, create a new aggregated transaction
				existingAggregatedTransaction = AggregatedTransaction{
					Category:      aggTransaction.Category,
					TotalQuantity: 0,
					TotalAmount:   0,
					Count:         0,
				}
			}

			// Update the aggregated transaction with the current transaction data
			existingAggregatedTransaction.TotalQuantity += aggTransaction.TotalQuantity
			existingAggregatedTransaction.TotalAmount += aggTransaction.TotalAmount
			existingAggregatedTransaction.Count += aggTransaction.Count

			// Store the aggregated transaction back into the map
			mergedAggregatedTransactions[aggTransaction.Category] = existingAggregatedTransaction
		}
		aggregatorWg.Done()
	}

	return mergedAggregatedTransactions
}

func ProcessTransactions(aggregatedTransactions interface{}) []ProcessedTransaction {
	processedTransactions := make([]ProcessedTransaction, 0)

	if aggregatedTransactionsSlice, ok := aggregatedTransactions.([]AggregatedTransaction); ok {
		for _, aggregatedTransaction := range aggregatedTransactionsSlice {
			processedTransactions = append(processedTransactions, processSingleTransaction(aggregatedTransaction))
		}
	}

	if aggregatedTransactionsMap, ok := aggregatedTransactions.(map[string]AggregatedTransaction); ok {
		for _, aggregatedTransaction := range aggregatedTransactionsMap {
			processedTransactions = append(processedTransactions, processSingleTransaction(aggregatedTransaction))
		}
	}

	return processedTransactions
}

func processSingleTransaction(aggregatedTransaction AggregatedTransaction) ProcessedTransaction {
	processedTransaction := ProcessedTransaction{
		Category:    "Processed_" + aggregatedTransaction.Category,
		TotalSales:  aggregatedTransaction.TotalAmount,
		AvgQuantity: float64(aggregatedTransaction.TotalQuantity) / float64(aggregatedTransaction.Count),
	}

	return processedTransaction
}
