package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func processTransactionsSlice(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	customerID := params.Get("customerID")
	withCache := params.Get("withCache")
	useCache := withCache == "true"
	prefix := params.Get("prefix")

	w.Header().Set("Content-Type", "application/json")

	ctx := context.Background()

	workingTransactionsSlice, respErr := getTransactionsSlice(ctx, useCache, customerID)
	var filteredTransactionsIDs []Transaction

	if params.Get("ptree") == "true" {
		// build the trie for performant naming prefix matching
		transactionsTrie := constructPrefixTree(workingTransactionsSlice)
		filteredTransactionsIDs = filterByPrefixTree(transactionsTrie, prefix)
	} else {
		filteredTransactionsIDs = simpleFilterByPrefix(workingTransactionsSlice, prefix)
	}

	if respErr != nil {
		fmt.Fprintf(w, respErr.Error())
	} else {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(filteredTransactionsIDs); err != nil {
			fmt.Fprintf(w, err.Error())
		}
	}
}

func simpleFilterByPrefix(transactions []Transaction, prefix string) []Transaction {
	filteredTransactions := make([]Transaction, 0)
	for _, transaction := range transactions {
		if strings.HasPrefix(transaction.ProductName, prefix) {
			filteredTransactions = append(filteredTransactions, transaction)
		}
	}
	return filteredTransactions
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

func findTransactionByID(transactions []Transaction, id string) Transaction {
	for _, tx := range transactions {
		if tx.TransactionID == id {
			return tx
		}
	}
	return Transaction{} // Return an empty transaction if not found
}
