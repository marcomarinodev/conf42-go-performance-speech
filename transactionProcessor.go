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

func main() {
	// initDb()
	fmt.Println("running server at 8080")
	http.HandleFunc("/transactions", requestHandler)
	http.ListenAndServe(":8080", nil)
}

func requestHandler(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	customerID := params.Get("customerID")

	// Parse the withCache parameter from the query params
	withCache := params.Get("withCache")
	useCache := withCache == "true" // Assuming "true" indicates using cache, otherwise only MongoDB

	w.Header().Set("Content-Type", "application/json")

	var respMessage map[string]interface{}
	var respErr error

	ctx := context.Background()

	if useCache {
		// Use both cache and MongoDB
		isCached, transactionsCache, err := getFromCache(ctx)
		if err != nil {
			respErr = err
		} else {
			if isCached {
				respMessage = transactionsCache
				respMessage["_source"] = "Redis Cache"
			} else {
				respMessage, err = getFromDb(ctx, customerID)
				if err != nil {
					respErr = err
				}
				err = addToCache(ctx, respMessage)
				if err != nil {
					respErr = err
				}
				respMessage["_source"] = "MongoDB database"
			}
		}
	} else {
		// Use only MongoDB
		respMessage, respErr = getFromDb(ctx, customerID)
		if respErr == nil {
			respMessage["_source"] = "MongoDB database"
		}
	}

	if respErr != nil {
		fmt.Fprintf(w, respErr.Error())
	} else {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(respMessage); err != nil {
			fmt.Fprintf(w, err.Error())
		}
	}
}

func getFromDb(ctx context.Context, customerId string) (map[string]interface{}, error) {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		return nil, err
	}

	collection := client.Database(MONGO_DB_NAME).Collection("transactions")
	filter := bson.D{{"customerID", customerId}}
	cur, err := collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var records []bson.M

	for cur.Next(ctx) {

		var record bson.M

		if err = cur.Decode(&record); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	res := map[string]interface{}{
		"data": records,
	}

	return res, nil
}

func getFromCache(ctx context.Context) (bool, map[string]interface{}, error) {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	transactionsCache, err := redisClient.Get("transactions_cache").Bytes()

	if err != nil {
		return false, nil, nil
	}

	res := map[string]interface{}{}

	err = json.Unmarshal(transactionsCache, &res)

	if err != nil {
		return false, nil, nil
	}

	return true, res, nil
}

func addToCache(ctx context.Context, data map[string]interface{}) error {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	jsonString, err := json.Marshal(data)

	if err != nil {
		return err
	}

	err = redisClient.Set("transactions_cache", jsonString, CACHE_EXP_TIME).Err()

	if err != nil {
		return nil
	}

	return nil
}
