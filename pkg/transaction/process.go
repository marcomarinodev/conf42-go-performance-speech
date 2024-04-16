package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func ProcessTransactionPipeline(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	customerID := params.Get("customerID")
	useCache := params.Get("withCache") == "true"
	prefix := params.Get("prefix")
	w.Header().Set("Content-Type", "application/json")
	ctx := context.Background()

	// Timing the total execution
	totalStartTime := time.Now()

	// Retrieval stage
	fmt.Println("Getting transactions slice...")
	workingTransactionsSlice, respErr := GetTransactionsSlice(ctx, useCache, customerID)
	if respErr != nil {
		fmt.Println(w, respErr.Error())
	}

	// Stage 1: Start Timer
	stage1StartTime := time.Now()
	res := StartPipeline_Seq(workingTransactionsSlice, prefix)
	stage1Duration := time.Since(stage1StartTime)
	fmt.Println("Stage 1 (StartPipeline_Seq) took:", stage1Duration)

	// Serialization stage
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		fmt.Println(w, err.Error())
	}
	fmt.Println("Total pipeline took:", time.Since(totalStartTime))
}

func ProcessTransactionPipeline_FirstOpt(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	customerID := params.Get("customerID")
	useCache := params.Get("withCache") == "true"
	prefix := params.Get("prefix")
	w.Header().Set("Content-Type", "application/json")
	ctx := context.Background()

	// Timing the total execution
	totalStartTime := time.Now()

	// Retrieval stage
	fmt.Println("Getting transactions slice...")
	workingTransactionsSlice, respErr := GetTransactionsSlice(ctx, useCache, customerID)
	if respErr != nil {
		fmt.Println(w, respErr.Error())
	}

	// Stage 1: Start Timer
	stage1StartTime := time.Now()
	res := StartPipeline_FirstOpt(workingTransactionsSlice, prefix, 4)
	stage1Duration := time.Since(stage1StartTime)
	fmt.Println("Stage 1 (StartPipeline_FirstOpt) took:", stage1Duration)

	// Serialization stage
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		fmt.Println(w, err.Error())
	}
	fmt.Println("Total pipeline took:", time.Since(totalStartTime))
}
