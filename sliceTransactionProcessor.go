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

var workingTransactionsSlice []Transaction

func processSliceTransactions(w http.ResponseWriter, req *http.Request) {
	// Calculate total revenue
	totalRevenue := calculateTotalRevenueSlice(workingTransactionsSlice)

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

func calculateTotalRevenueSlice(transactions []Transaction) float64 {
	totalRevenue := 0.0
	// Iterate over each transaction in the slice
	for _, transaction := range transactions {
		// Add the total amount of each transaction to the total revenue
		totalRevenue += transaction.TotalAmount
	}
	return totalRevenue
}

func requestSliceHandler(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	customerID := params.Get("customerID")

	// Parse the withCache parameter from the query params
	withCache := params.Get("withCache")
	useCache := withCache == "true" // Assuming "true" indicates using cache, otherwise only MongoDB

	w.Header().Set("Content-Type", "application/json")

	var respErr error

	ctx := context.Background()

	if useCache {
		isCached, cacheTransactions, err := getSliceFromCache(customerID)
		workingTransactionsSlice = cacheTransactions
		if err != nil {
			respErr = err
		} else {
			if !isCached {
				workingTransactionsSlice, err = getSliceFromDb(ctx, customerID)
				if err != nil {
					respErr = err
				}

				err = addsSliceToCache(customerID, workingTransactionsSlice)
				if err != nil {
					respErr = err
				}
			}
		}
	} else {
		// Use only MongoDB
		workingTransactionsSlice, respErr = getSliceFromDb(ctx, customerID)
	}

	if respErr != nil {
		fmt.Fprintf(w, respErr.Error())
	} else {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(workingTransactionsSlice); err != nil {
			fmt.Fprintf(w, err.Error())
		}
	}
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

func findTransactionByID(transactions []Transaction, id string) Transaction {
	for _, tx := range transactions {
		if tx.TransactionID == id {
			return tx
		}
	}
	return Transaction{} // Return an empty transaction if not found
}
