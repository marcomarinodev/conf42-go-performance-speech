package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var workingTransactionsMap map[string]Transaction

func processMapTransactions(w http.ResponseWriter, req *http.Request) {
	// Calculate total revenue
	totalRevenue := calculateTotalRevenueMap(workingTransactionsMap)

	// Convert totalRevenue to JSON format
	response, err := json.Marshal(map[string]float64{"total_revenue": totalRevenue})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")
	// Write the response
	w.Write(response)
}

func calculateTotalRevenueMap(transactions map[string]Transaction) float64 {
	totalRevenue := 0.0
	// Iterate over each transaction in the map
	for _, transaction := range transactions {
		// Add the total amount of each transaction to the total revenue
		totalRevenue += transaction.TotalAmount
	}
	return totalRevenue
}

func requestMapHandler(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	customerID := params.Get("customerID")
	withCache := params.Get("withCache")
	useCache := withCache == "true"

	w.Header().Set("Content-Type", "application/json")

	var respErr error

	ctx := context.Background()

	if useCache {
		isCached, cacheTransactions, err := getMapFromCache(customerID)
		workingTransactionsMap = cacheTransactions
		if err != nil {
			respErr = err
		} else {
			if !isCached {
				workingTransactionsMap, err = getMapFromDb(ctx, customerID)
				if err != nil {
					respErr = err
				}

				err = addMapToCache(customerID, workingTransactionsMap)
				if err != nil {
					respErr = err
				}
			}
		}
	} else {
		workingTransactionsMap, respErr = getMapFromDb(ctx, customerID)
	}

	if respErr != nil {
		fmt.Fprintf(w, respErr.Error())
	} else {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(workingTransactionsMap); err != nil {
			fmt.Fprintf(w, err.Error())
		}
	}
}

func getMapFromDb(ctx context.Context, customerID string) (map[string]Transaction, error) {

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

	records := make(map[string]Transaction)

	for cur.Next(ctx) {

		var currentRecord Transaction

		if err = cur.Decode(&currentRecord); err != nil {
			return nil, err
		}

		records[currentRecord.TransactionID] = currentRecord
	}

	return records, nil
}

func getMapFromCache(customerID string) (bool, map[string]Transaction, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	transactionsCache, err := redisClient.Get(customerID).Bytes()

	if err != nil {
		return false, nil, nil
	}

	res := make(map[string]Transaction)

	err = json.Unmarshal(transactionsCache, &res)
	if err != nil {
		return false, nil, nil
	}

	return true, res, nil
}

func addMapToCache(customerID string, transactions map[string]Transaction) error {
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
