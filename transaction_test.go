package main

import (
	"reflect"
	"testing"
)

func TestAggregateTransactions(t *testing.T) {
	// Sample transactions
	transactions := []Transaction{
		{Category: "A", Quantity: 2, TotalAmount: 20.0},
		{Category: "B", Quantity: 3, TotalAmount: 30.0},
		{Category: "A", Quantity: 1, TotalAmount: 10.0},
		{Category: "B", Quantity: 4, TotalAmount: 40.0},
		{Category: "C", Quantity: 5, TotalAmount: 50.0},
	}

	expected := []AggregatedTransaction{
		{Category: "A", TotalQuantity: 3, TotalAmount: 30.0, Count: 2},
		{Category: "B", TotalQuantity: 7, TotalAmount: 70.0, Count: 2},
		{Category: "C", TotalQuantity: 5, TotalAmount: 50.0, Count: 1},
	}

	actual := aggregateTransactions(transactions)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestProcessTransactions(t *testing.T) {
	// Sample aggregated transactions
	aggregatedTransactions := []AggregatedTransaction{
		{Category: "A", TotalQuantity: 3, TotalAmount: 30.0, Count: 2},
		{Category: "B", TotalQuantity: 7, TotalAmount: 70.0, Count: 2},
		{Category: "C", TotalQuantity: 5, TotalAmount: 50.0, Count: 1},
	}

	// Expected processed transactions
	expected := []ProcessedTransaction{
		{Category: "Processed_A", TotalSales: 30.0, AvgQuantity: 1.5},
		{Category: "Processed_B", TotalSales: 70.0, AvgQuantity: 3.5},
		{Category: "Processed_C", TotalSales: 50.0, AvgQuantity: 5.0},
	}

	actual := processTransactions(aggregatedTransactions)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
