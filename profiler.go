package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime/pprof"
)

func startCPUProfile(w http.ResponseWriter, r *http.Request) {
	f, err := os.Create("cpu.pprof")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "CPU profiling started")
}

func stopCPUProfile(w http.ResponseWriter, r *http.Request) {
	pprof.StopCPUProfile()
	fmt.Fprintln(w, "CPU profiling stopped")
}
