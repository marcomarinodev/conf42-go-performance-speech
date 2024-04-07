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
	http.HandleFunc("/getSliceTransactions", requestSliceHandler)
	http.HandleFunc("/getMapTransactions", requestMapHandler)
	http.HandleFunc("/processSliceTransactions", processSliceTransactions)
	http.HandleFunc("/processMapTransactions", processMapTransactions)
	http.HandleFunc("/printSliceTransactions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(workingTransactionsSlice)
	})
	http.HandleFunc("/printMapTransactions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(workingTransactionsMap)
	})
	http.ListenAndServe(":8080", nil)
}
