package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// write behind caching: the application reads and writes data to Redis
// redis syncs any changed data to the db asynchronously
// that's because a service can only add new transactions and not modifying or deleting
// existing ones. Since services reads only from Redis, the data is always up to date.

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func getTransactionsByCustomerIDFromRedis(customerID string, cacheFlag bool) ([]Transaction, error) {

	transactions := make([]Transaction, 0)

	if cacheFlag {
		// TODO use a smarter get as the cache entry key is the transaction ID and not customer ID
		// check if data is cached in Redis
		keys, err := rdb.Keys(ctx, "*").Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			cachedTransaction, err := rdb.Get(ctx, key).Result()
			if err == nil {
				var transaction Transaction
				err = json.Unmarshal([]byte(cachedTransaction), &transaction)
				if err == nil && transaction.CustomerID == customerID {
					transactions = append(transactions, transaction)
				}
			}
		}
	}

	if cacheFlag {
		// Cache data in Redis for future use
		transactionsJSON, err := json.Marshal(transactions)
		if err != nil {
			return nil, err
		}

		fmt.Println("saving the transactions into the cache")

		// Set the JSON data with a specific key
		// rdb.Set(ctx, customerID, string(transactionsJSON), 0)

		// Set the data as a JSON object, use transaction ID as a key
		rdb.JSONSet(ctx, customerID, "$", string(transactionsJSON))
	}

	return transactions, nil
}

func getTransactionsByCustomerIDFromDB(customerID string) ([]Transaction, error) {
	transactions := make([]Transaction, 0)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database("ecommerce").Collection("transactions")

	var cursor *mongo.Cursor
	var cursorErr error

	if customerID == "" {
		cursor, cursorErr = collection.Find(context.TODO(), bson.M{})
	} else {
		cursor, cursorErr = collection.Find(context.TODO(), bson.M{"customerID": customerID})
	}

	if cursorErr != nil {
		return nil, cursorErr
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var transaction Transaction
		if err := cursor.Decode(&transaction); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// TODO this function is useful to proof the write behind caching
// func editTransaction(transactionID string, newCustomerID string) error {
// 	// get the transaction from the cache
// 	transaction := rdb.JSONGet(ctx, transactionID, "$")

// 	// update the transaction
// 	err := rdb.JSONSet(ctx, transactionID, "$.customerID", newCustomerID)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func main() {
	initDbFlag := flag.Bool("initDb", false, "Initialize the database")
	cacheFlag := flag.Bool("cache", false, "Cache transactions")
	flag.Parse()

	if *initDbFlag {
		initDb()
	}

	http.HandleFunc("/allTransactions", func(w http.ResponseWriter, r *http.Request) {
		transactions, err := getTransactionsByCustomerID("", *cacheFlag)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	})

	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		customerID := r.URL.Query().Get("customerID")
		if customerID == "" {
			http.Error(w, "Missing customerID parameter", http.StatusBadRequest)
			return
		}

		transactions, err := getTransactionsByCustomerID(customerID, *cacheFlag)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	})

	http.HandleFunc("/start-cpu-profile", startCPUProfile)
	http.HandleFunc("/stop-cpu-profile", stopCPUProfile)

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
