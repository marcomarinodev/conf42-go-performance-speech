package tests

import (
	"testing"
	"transactions_sample/pkg/transaction"
)

var y map[string]transaction.AggregatedTransaction
var y2 []transaction.AggregatedTransaction

func BenchmarkCustomPipeline_Seq(b *testing.B) {
	allTransactions := transaction.GenerateTransactionsForTest(2000000)
	prefix := "US"
	b.ResetTimer()
	var x []transaction.AggregatedTransaction
	for i := 0; i < b.N; i++ {
		x = transaction.StartPipeline_Seq(allTransactions, prefix)
	}
	y2 = x
}

func BenchmarkCustomPipeline_FirstOpt(b *testing.B) {
	allTransactions := transaction.GenerateTransactionsForTest(2000000)
	prefix := "U"
	numWorkers := 20
	b.ResetTimer()
	var x map[string]transaction.AggregatedTransaction
	for i := 0; i < b.N; i++ {
		x = transaction.StartPipeline_FirstOpt(allTransactions, prefix, numWorkers)
	}

	y = x
}

func BenchmarkCustomPipeline_Optimized_SecOpt(b *testing.B) {
	allTransactions := transaction.GenerateTransactionsForTest(2000000)
	prefix := "U"
	var x map[string]transaction.AggregatedTransaction
	for i := 0; i < b.N; i++ {
		x = transaction.StartPipeline(&allTransactions, prefix, 4, 4)
	}

	y = x
}

// ====================================================
// ==================== TESTS =========================
// ====================================================
func TestCustomPipelineFirstOpt(t *testing.T) {
	allTransactions := transaction.GenerateTransactionsForTest(10000 * 100)
	prefix := "US"
	numWorkers := 4

	expected := transaction.StartPipeline_Seq(allTransactions, prefix)
	actual := transaction.StartPipeline_FirstOpt(allTransactions, prefix, numWorkers)

	if !EqualAggregatedTransactions(expected, actual) {
		t.Errorf("Transactions slices are not equal")
	}

}

func TestCustomPipelineSecOpt(t *testing.T) {
	allTransactions := transaction.GenerateTransactionsForTest(10000 * 100)
	prefix := "US"

	expected := transaction.StartPipeline_Seq(allTransactions, prefix)
	actual := transaction.StartPipeline(&allTransactions, prefix, 4, 4)

	if !EqualAggregatedTransactions(expected, actual) {
		t.Errorf("Transactions slices are not equal")
	}
}
