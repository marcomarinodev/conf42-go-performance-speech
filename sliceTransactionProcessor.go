package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func processTransactionsSlice(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	customerID := params.Get("customerID")
	useCache := params.Get("withCache") == "true"
	prefix := params.Get("prefix")
	w.Header().Set("Content-Type", "application/json")
	ctx := context.Background()

	// Retrieval stage
	fmt.Println("Getting transactions slice...")
	workingTransactionsSlice, respErr := getTransactionsSlice(ctx, useCache, customerID)
	if respErr != nil {
		fmt.Println(w, respErr.Error())
	}

	totalStartTime := time.Now()

	// Filtering stage
	fmt.Println("Filtering...")
	filteringStartTime := time.Now()
	filteredTransactions := filterByPrefix_seq(workingTransactionsSlice, prefix)
	fmt.Println("Filtering took: " + time.Since(filteringStartTime).String())

	// Aggregation stage
	fmt.Println("Aggregating...")
	aggregationStartTime := time.Now()
	res := aggregateTransactions_seq(filteredTransactions)
	fmt.Println("Aggregating took: " + time.Since(aggregationStartTime).String())

	// Processing stage
	// processingStartTime := time.Now()
	// processedTransactions := processTransactions(aggregatedTransactions)
	// fmt.Println("Aggregating took: " + time.Since(processingStartTime).String())

	for _, aggRes := range res {
		fmt.Println(aggRes)
	}

	// Serialization stage
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode("ok"); err != nil {
		fmt.Println(w, err.Error())
	}
	fmt.Println("Total pipeline took: " + time.Since(totalStartTime).String())
}

func processTransactionsSlice_Optimized(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	customerID := params.Get("customerID")
	useCache := params.Get("withCache") == "true"
	prefix := params.Get("prefix")
	w.Header().Set("Content-Type", "application/json")
	ctx := context.Background()

	// Retrieval stage
	workingTransactionsSlice, respErr := getTransactionsSlice(ctx, useCache, customerID)
	if respErr != nil {
		fmt.Println(w, respErr.Error())
	}

	totalStartTime := time.Now()

	// Filtering stage
	fmt.Println("Filtering...")
	filteringStartTime := time.Now()
	filterTransactionsChan := filterByPrefix_par(workingTransactionsSlice, prefix, 4)
	fmt.Println("Filtering took: " + time.Since(filteringStartTime).String())

	// Aggregation stage
	fmt.Println("Aggregating...")
	aggregationStartTime := time.Now()

	aggResultChan := make(chan []AggregatedTransaction)
	aggNumWorkers := 4

	// fan-out: create worker goroutines
	for i := 0; i < aggNumWorkers; i++ {
		go aggregatorWorker(i, filterTransactionsChan, aggResultChan)
	}

	// fan-in: collect results
	var aggWg sync.WaitGroup
	aggWg.Add(aggNumWorkers)

	go func() {
		aggWg.Wait()         // Wait for all jobs to be done
		close(aggResultChan) // Close the results channel after all jobs are processed
	}()

	res := mergeAggregatedTransactions(aggResultChan, &aggWg)

	fmt.Println("Aggregating took: " + time.Since(aggregationStartTime).String())

	for _, aggResult := range res {
		fmt.Println(aggResult)
	}

	// Processing stage
	// processingStartTime := time.Now()
	// processedTransactions := processTransactions(aggregatedTransactions)
	// fmt.Println("Aggregating took: " + time.Since(processingStartTime).String())

	// Serialization stage
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode("ok"); err != nil {
		fmt.Println(w, err.Error())
	}
	fmt.Println("Total pipeline took: " + time.Since(totalStartTime).String())
}

func getTransactionsSlice(ctx context.Context, useCache bool, customerID string) ([]Transaction, error) {
	var slice []Transaction
	if useCache {
		isCached, cacheTransactions, err := getSliceFromCache(customerID)
		slice = cacheTransactions
		if err != nil {
			return nil, err
		} else {
			if !isCached {
				slice, err = getSliceFromDb(ctx, customerID)
				if err != nil {
					return nil, err
				}

				err = addsSliceToCache(customerID, slice)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		var err error
		slice, err = getSliceFromDb(ctx, customerID)
		if err != nil {
			return nil, err
		}
	}
	return slice, nil
}

func getSliceFromDb(ctx context.Context, customerID string) ([]Transaction, error) {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		return nil, err
	}

	collection := client.Database(MONGO_DB_NAME).Collection("transactions")
	filter := bson.D{{"customerID", customerID}}
	cur, err := collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var records []Transaction

	for cur.Next(ctx) {

		var currentRecord Transaction

		if err = cur.Decode(&currentRecord); err != nil {
			return nil, err
		}

		records = append(records, currentRecord)
	}

	return records, nil
}

func getSliceFromCache(customerID string) (bool, []Transaction, error) {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	transactionsCache, err := redisClient.Get(customerID).Bytes()

	if err != nil {
		return false, nil, nil
	}

	var res []Transaction

	err = json.Unmarshal(transactionsCache, &res)

	if err != nil {
		return false, nil, nil
	}

	return true, res, nil
}

func addsSliceToCache(customerID string, transactions []Transaction) error {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	jsonString, err := json.Marshal(transactions)

	if err != nil {
		return err
	}

	err = redisClient.Set(customerID, jsonString, CACHE_EXP_TIME).Err()

	if err != nil {
		return nil
	}

	return nil
}
