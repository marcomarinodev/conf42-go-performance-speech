package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	MONGO_DB_NAME  = "store"
	CACHE_EXP_TIME = 30 * time.Second
)

func main() {
	// initDb()
	fmt.Println("running server at 8080")
	http.HandleFunc("/processSliceTransactions", processTransactionsSlice)
	http.HandleFunc("/processSliceTransactionsOptimize", processTransactionsSlice_Optimized)
	// http.HandleFunc("/getMapTransactions", requestMapHandler)

	http.ListenAndServe(":8080", nil)
}
