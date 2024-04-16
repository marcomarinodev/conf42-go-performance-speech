package transaction

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

func StartPipeline_Seq(allTransactions []Transaction, prefix string) []AggregatedTransaction {
	filteredTransactions := filterByPrefix_seq(allTransactions, prefix)
	return aggregateTransactions_seq(filteredTransactions)
}

func StartPipeline_FirstOpt(allTransactions []Transaction, prefix string, numWorkers int) map[string]AggregatedTransaction {
	var wg sync.WaitGroup
	aggregatorsMap := sync.Map{}
	partSize := len(allTransactions) / numWorkers

	for workerID := 0; workerID < numWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			filteredTransactions := filterByPrefix_seq(allTransactions[id*partSize:(id+1)*partSize], prefix)
			aggregatorsMap.Store(id, aggregateTransactions_seq(filteredTransactions))
		}(workerID)
	}

	wg.Wait() // Wait for all goroutines to finish

	var actualAggregateResult []AggregatedTransaction

	for i := range numWorkers {
		subAggregatedTransactions, _ := aggregatorsMap.Load(i)
		actualAggregateResult = append(actualAggregateResult, subAggregatedTransactions.([]AggregatedTransaction)...)
	}

	return mergeAggregatedTransactions(actualAggregateResult)
}

func aggregateWorker(
	id int,
	inputAggregateChan <-chan Transaction,
	aggWg *sync.WaitGroup,
	aggregatorsMap map[int]map[string]AggregatedTransaction,
	// aggregatorsMapMtx *sync.Mutex,
) {
	defer aggWg.Done()
	for t := range inputAggregateChan {
		aggTransaction, ok := aggregatorsMap[id][t.Category]
		if !ok {
			// If the category doesn't exist, create a new aggregated transaction
			aggTransaction = AggregatedTransaction{
				Category:      t.Category,
				TotalQuantity: 0,
				TotalAmount:   0,
				Count:         0,
			}
		}

		// Update the aggregated transaction with the current transaction data
		aggTransaction.Category = t.Category
		aggTransaction.TotalQuantity += t.Quantity
		aggTransaction.TotalAmount += t.TotalAmount
		aggTransaction.Count++

		// aggregatorsMapMtx.Lock()
		// Store the aggregated transaction back into the map
		aggregatorsMap[id][t.Category] = aggTransaction
		// aggregatorsMapMtx.Unlock()
	}
}

func filterWorker(
	prefix string,
	inputFilterSlice *[]Transaction,
	mtx *sync.Mutex,
	inputAggregateChan chan<- Transaction,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		mtx.Lock()
		if len(*inputFilterSlice) == 0 {
			mtx.Unlock()
			return
		}

		index := len(*inputFilterSlice) - 1
		t := (*inputFilterSlice)[index]
		*inputFilterSlice = (*inputFilterSlice)[:index]
		mtx.Unlock()

		if strings.HasPrefix(t.ProductName, prefix) {
			inputAggregateChan <- t
		}
	}
}

func filterByPrefix_seq(transactions []Transaction, prefix string) []Transaction {
	var filteredTransactions []Transaction
	for _, transaction := range transactions {
		if strings.HasPrefix(transaction.ProductName, prefix) {
			filteredTransactions = append(filteredTransactions, transaction)
		}
	}

	return filteredTransactions
}

func aggregateTransactions_seq(transactions []Transaction) []AggregatedTransaction {
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

func mergeAggregatedTransactions(aggTransactions []AggregatedTransaction) map[string]AggregatedTransaction {
	mergedAggregatedTransactions := make(map[string]AggregatedTransaction)

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

	return mergedAggregatedTransactions
}
