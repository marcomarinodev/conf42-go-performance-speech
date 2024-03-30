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

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func getTransactionsByCustomerID(customerID string, cacheFlag bool) ([]Transaction, error) {

	if cacheFlag {
		// check if data is cached in Redis
		cachedTransactions, err := rdb.Get(ctx, customerID).Result()
		if err == nil {
			var transactions []Transaction
			err = json.Unmarshal([]byte(cachedTransactions), &transactions)
			if err == nil {
				return transactions, nil
			}

			return nil, err
		}
	}

	// if data is not cached, fetch from mongo
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database("ecommerce").Collection("transactions")

	var transactions []Transaction
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

	if cacheFlag {
		// Cache data in Redis for future use
		transactionsJSON, err := json.Marshal(transactions)
		if err != nil {
			return nil, err
		}

		rdb.Set(ctx, customerID, string(transactionsJSON), 0)
	}

	return transactions, nil
}

func main() {
	initDbFlag := flag.Bool("init-db", false, "Initialize the database")
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
