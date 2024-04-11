package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/beevik/prefixtree"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var transactionsPtree *prefixtree.Tree

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

	// Filtering stage
	fmt.Println("Filtering...")
	now := time.Now()
	filteredTransactionsIDs := simpleFilterByPrefixFromSlice(workingTransactionsSlice, prefix)
	elapsed := time.Since(now)
	fmt.Println("Finished sequential processing transactions slice: " + elapsed.String())

	// Aggregation stage
	aggregatedTransactions := aggregateTransactions(filteredTransactionsIDs)

	// Processing stage
	processedTransactions := processTransactions(aggregatedTransactions)

	// Serialization stage
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(processedTransactions); err != nil {
		fmt.Println(w, err.Error())
	}
}

func processTransactionsSlice_Optimized(w http.ResponseWriter, req *http.Request) {
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

	// Filtering stage
	fmt.Println("Filtering...")
	now := time.Now()
	t := parallelFilterByPrefixFromSlice(workingTransactionsSlice, prefix, 4)
	elapsed := time.Since(now)
	fmt.Println("Finished parallel processing transactions slice: " + elapsed.String())
	// Aggregation stage
	// aggregationResult := make(chan []AggregatedTransaction)
	// go parallelAggregateTransactions(filteringResult, aggregationResult)

	// // Wait for aggregation to complete
	// aggregatedTransactions := <-aggregationResult
	// Processing stage
	// processedTransactions := processTransactions(aggregatedTransactions)

	// Serialization stage
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(t); err != nil {
		fmt.Println(w, err.Error())
	}
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
