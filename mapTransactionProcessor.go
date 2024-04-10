package main

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// func processTransactionsMap(w http.ResponseWriter, req *http.Request) {
// 	params := req.URL.Query()
// 	customerID := params.Get("customerID")
// 	withCache := params.Get("withCache")
// 	useCache := withCache == "true"
// 	prefix := params.Get("prefix")

// 	w.Header().Set("Content-Type", "application/json")

// 	ctx := context.Background()

// 	workingTransactionsMap, respErr := getTransactionsMap(ctx, useCache, customerID)
// 	var filteredTransactionsIDs []Transaction

// 	if params.Get("optimize") == "true" {
// 		// build the trie for performant naming prefix matching
// 		transactionsTrie := constructTrie(workingTransactionsSlice)
// 		filteredTransactionsIDs = trieFilterByPrefix(transactionsTrie, prefix)
// 	} else {
// 		filteredTransactionsIDs = simpleFilterByPrefix(workingTransactionsSlice, prefix)
// 	}

// 	if respErr != nil {
// 		fmt.Fprintf(w, respErr.Error())
// 	} else {
// 		enc := json.NewEncoder(w)
// 		enc.SetIndent("", "  ")
// 		if err := enc.Encode(filteredTransactionsIDs); err != nil {
// 			fmt.Fprintf(w, err.Error())
// 		}
// 	}
// }

func getTransactionsMap(ctx context.Context, useCache bool, customerID string) (map[string]Transaction, error) {
	var transactionsMap map[string]Transaction
	var err error
	var isCached bool
	if useCache {
		isCached, transactionsMap, err = getMapFromCache(customerID)
		if err != nil {
			return nil, err
		} else {
			if !isCached {
				transactionsMap, err = getMapFromDb(ctx, customerID)
				if err != nil {
					return nil, err
				}

				err = addMapToCache(customerID, transactionsMap)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		var err error
		transactionsMap, err = getMapFromDb(ctx, customerID)
		if err != nil {
			return nil, err
		}
	}
	return transactionsMap, nil
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
