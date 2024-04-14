package tests

import (
	"math"
	"reflect"
	"transactions_sample/pkg/transaction"
)

func EqualTransactions(slice1, slice2 []transaction.Transaction) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	// Create maps to count occurrences of each transaction in both slices
	count1 := make(map[transaction.Transaction]int)
	count2 := make(map[transaction.Transaction]int)

	for _, t := range slice1 {
		count1[t]++
	}
	for _, t := range slice2 {
		count2[t]++
	}

	// Compare the counts of transactions in both maps
	return reflect.DeepEqual(count1, count2)
}

func EqualAggregatedTransactions(sliceT []transaction.AggregatedTransaction, mapT map[string]transaction.AggregatedTransaction) bool {
	if len(sliceT) != len(mapT) {
		return false
	}

	// Create maps to count occurrences of each transaction in both slices
	count1 := make(map[transaction.AggregatedTransaction]int)
	count2 := make(map[transaction.AggregatedTransaction]int)

	for _, t := range sliceT {
		roundedT := t
		roundedT.TotalAmount = reduceToTwoDigits(roundedT.TotalAmount)
		count1[roundedT]++
	}
	for _, t := range mapT {
		roundedT := t
		roundedT.TotalAmount = reduceToTwoDigits(roundedT.TotalAmount)
		count2[roundedT]++
	}

	// Compare the counts of transactions in both maps
	return reflect.DeepEqual(count1, count2)
}

func reduceToTwoDigits(num float64) float64 {
	return math.Round(num*100) / 100
}
