package main

import (
	"fmt"
	"net/http"
	"transactions_sample/pkg/transaction"
)

func main() {
	// initDb()
	fmt.Println("running server at 8080")
	http.HandleFunc("/seq", transaction.ProcessTransactionPipeline)
	http.HandleFunc("/first_opt", transaction.ProcessTransactionPipeline_FirstOpt)
	http.ListenAndServe(":8080", nil)
}
