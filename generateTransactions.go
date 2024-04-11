package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func initDb() {
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.TODO())

	// Select database and collection
	collection := client.Database("store").Collection("transactions")

	// Define slices for randomization
	prefixes := []string{"Wireless", "USB", "Gaming", "Portable", "Smart", "Professional", "High-Speed"}
	suffixes := []string{"Mouse", "Keyboard", "Cable", "Monitor", "Webcam", "Laptop", "Headphones", "Smartphone", "Tablet", "Printer"}
	categories := []string{"Electronics", "Computers", "Accessories", "Office"}
	unitPrices := []float64{29.99, 49.99, 5.99, 199.99, 89.99, 999.99, 99.99, 799.99, 299.99, 149.99}
	paymentMethods := []string{"Credit Card", "PayPal", "Debit Card", "Bitcoin"}

	// Generate and insert transactions
	fmt.Println("Generating transactions...")
	for i := 0; i < 500000; i++ {
		prefixIndex := rand.Intn(len(prefixes))
		suffixIndex := rand.Intn(len(suffixes))
		categoryIndex := rand.Intn(len(categories))
		priceIndex := rand.Intn(len(unitPrices))
		paymentMethodIndex := rand.Intn(len(paymentMethods))

		quantity := rand.Intn(5) + 1
		unitPrice := unitPrices[priceIndex]
		totalAmount := float64(quantity) * unitPrice

		transaction := Transaction{
			TransactionID: fmt.Sprintf("TXN%d", i),
			Timestamp:     time.Now(),
			CustomerID:    fmt.Sprintf("CUST%d", rand.Intn(1)),
			ProductName:   fmt.Sprintf("%s %s", prefixes[prefixIndex], suffixes[suffixIndex]),
			Category:      categories[categoryIndex],
			Quantity:      quantity,
			UnitPrice:     unitPrice,
			TotalAmount:   totalAmount,
			PaymentMethod: paymentMethods[paymentMethodIndex],
		}

		_, err := collection.InsertOne(context.TODO(), transaction)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Transactions generated and inserted into MongoDB")
}

func generateTransactionsForTest(size int) []Transaction {
	transactions := make([]Transaction, 0)
	// Define slices for randomization
	productNames := []string{"Wireless Mouse", "Keyboard", "USB Cable", "Monitor", "Webcam"}
	categories := []string{"Electronics", "Computers", "Accessories", "Office"}
	unitPrices := []float64{29.99, 49.99, 5.99, 199.99, 89.99}
	paymentMethods := []string{"Credit Card", "PayPal", "Debit Card", "Bitcoin"}

	// Generate and insert transactions
	for i := 0; i < size; i++ {
		productIndex := rand.Intn(len(productNames))
		categoryIndex := rand.Intn(len(categories))
		priceIndex := rand.Intn(len(unitPrices))
		paymentMethodIndex := rand.Intn(len(paymentMethods))

		quantity := rand.Intn(5) + 1
		unitPrice := unitPrices[priceIndex]
		totalAmount := float64(quantity) * unitPrice

		transaction := Transaction{
			TransactionID: fmt.Sprintf("TXN%d", i),
			Timestamp:     time.Now(),
			CustomerID:    "CUSTX",
			ProductName:   productNames[productIndex],
			Category:      categories[categoryIndex],
			Quantity:      quantity,
			UnitPrice:     unitPrice,
			TotalAmount:   totalAmount,
			PaymentMethod: paymentMethods[paymentMethodIndex],
		}

		transactions = append(transactions, transaction)
	}

	return transactions
}
