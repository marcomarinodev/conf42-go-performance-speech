package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Movie struct {
	MovieID       string    `bson:"movieID"`
	Timestamp     time.Time `bson:"timestamp"`
	CustomerID    string    `bson:"customerID"`
	ProductName   string    `bson:"productName"`
	Category      string    `bson:"category"`
	Quantity      int       `bson:"quantity"`
	UnitPrice     float64   `bson:"unitPrice"`
	TotalAmount   float64   `bson:"totalAmount"`
	PaymentMethod string    `bson:"paymentMethod"`
}

func initDb() {
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.TODO())

	// Select database and collection
	collection := client.Database("ecommerce").Collection("transactions")

	// Define slices for randomization with movies related stuff
	movieTitles := []string{"Inception", "The Matrix", "Interstellar", "The Dark Knight", "Pulp Fiction"}
	genres := []string{"Action", "Sci-Fi", "Drama", "Thriller"}
	prices := []float64{9.99, 12.99, 7.99, 14.99, 10.99}
	paymentMethods := []string{"Credit Card", "PayPal", "Debit Card", "Bitcoin"}

	// Generate and insert movies
	for i := 0; i < 50000; i++ {
		titleIndex := rand.Intn(len(movieTitles))
		genreIndex := rand.Intn(len(genres))
		priceIndex := rand.Intn(len(prices))
		paymentMethodIndex := rand.Intn(len(paymentMethods))

		quantity := rand.Intn(5) + 1
		unitPrice := prices[priceIndex]
		totalAmount := float64(quantity) * unitPrice

		movie := Movie{
			MovieID:       fmt.Sprintf("MOV%d", rand.Intn(100000)),
			Timestamp:     time.Now(),
			CustomerID:    fmt.Sprintf("CUST%d", rand.Intn(30)),
			ProductName:   movieTitles[titleIndex],
			Category:      genres[genreIndex],
			Quantity:      quantity,
			UnitPrice:     unitPrice,
			TotalAmount:   totalAmount,
			PaymentMethod: paymentMethods[paymentMethodIndex],
		}

		_, err := collection.InsertOne(context.TODO(), movie)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Movies generated and inserted into MongoDB")
}
