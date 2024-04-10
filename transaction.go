package main

import (
	"time"

	"github.com/beevik/prefixtree"
)

type Transaction struct {
	TransactionID string    `bson:"transactionID"`
	Timestamp     time.Time `bson:"timestamp"`
	CustomerID    string    `bson:"customerID"`
	ProductName   string    `bson:"productName"`
	Category      string    `bson:"category"`
	Quantity      int       `bson:"quantity"`
	UnitPrice     float64   `bson:"unitPrice"`
	TotalAmount   float64   `bson:"totalAmount"`
	PaymentMethod string    `bson:"paymentMethod"`
}

func constructPrefixTree(transactions []Transaction) *prefixtree.Tree {
	tree := prefixtree.New()
	for _, transaction := range transactions {
		ts, err := tree.FindValue(transaction.ProductName)
		if err != nil {
			newTransactionSlice := make([]Transaction, 0)
			newTransactionSlice = append(newTransactionSlice, transaction)
			tree.Add(transaction.ProductName, newTransactionSlice)
			continue
		}

		existingTransactions := ts.([]Transaction)
		existingTransactions = append(existingTransactions, transaction)
		tree.Add(transaction.ProductName, existingTransactions)
	}

	return tree
}

func filterByPrefixTree(trie *prefixtree.Tree, prefix string) []Transaction {
	filteredTransactionsAny := trie.FindValues(prefix)
	filteredTransactions := make([]Transaction, 0)

	for i := 0; i < len(filteredTransactionsAny); i++ {
		value := filteredTransactionsAny[i].([]Transaction)
		filteredTransactions = append(filteredTransactions, value...)
	}

	return filteredTransactions
}
