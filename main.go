package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	MONGO_DB_NAME  = "store"
	CACHE_EXP_TIME = 30 * time.Minute
)

func main() {
	// initDb()
	fmt.Println("running server at 8080")
	http.HandleFunc("/processSliceTransactions", processTransactionPipeline)
	http.HandleFunc("/processSliceTransactionsOptimize", processTransactionPipeline_Optimized)

	http.ListenAndServe(":8080", nil)
}
