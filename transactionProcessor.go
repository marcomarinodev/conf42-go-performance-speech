package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MONGO_DB_NAME  = "store"
	CACHE_EXP_TIME = 30 * time.Second
)

var workingTransactions []Transaction

func requestHandler(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	customerID := params.Get("customerID")

	// Parse the withCache parameter from the query params
	withCache := params.Get("withCache")
	useCache := withCache == "true" // Assuming "true" indicates using cache, otherwise only MongoDB

	w.Header().Set("Content-Type", "application/json")

	var respErr error
	var transactions []Transaction

	ctx := context.Background()

	if useCache {
		// Use both cache and MongoDB
		isCached, cacheTransactions, err := getFromCache(customerID)
		transactions = cacheTransactions
		if err != nil {
			respErr = err
		} else {
			if !isCached {
				transactions, err = getFromDb(ctx, customerID)
				if err != nil {
					respErr = err
				}

				err = addToCache(customerID, transactions)
				if err != nil {
					respErr = err
				}
			}
		}
	} else {
		// Use only MongoDB
		transactions, respErr = getFromDb(ctx, customerID)
	}

	if respErr != nil {
		fmt.Fprintf(w, respErr.Error())
	} else {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(transactions); err != nil {
			fmt.Fprintf(w, err.Error())
		}
		workingTransactions = transactions
	}
}

func getFromDb(ctx context.Context, customerID string) ([]Transaction, error) {

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

func getFromCache(customerID string) (bool, []Transaction, error) {

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

func addToCache(customerID string, transactions []Transaction) error {

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

func main() {
	// initDb()
	fmt.Println("running server at 8080")
	http.HandleFunc("/transactions", requestHandler)
	http.HandleFunc("/printTransactions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(workingTransactions)
	})
	http.ListenAndServe(":8080", nil)
}
