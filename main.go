package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
)

func healthcheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func runPipelineHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	RunPipeline1(context.Background(), []string{"FOO", "BAR", "BAX", "", "XYZ"})
	fmt.Fprintln(w, "OK")
}

func registerPProfHandlers(r *http.ServeMux) {
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func main() {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", healthcheckHandler)
	m.HandleFunc("/run", runPipelineHandler)
	registerPProfHandlers(m)
	if err := http.ListenAndServe(":8080", m); err != nil {
		log.Fatal(err)
	}
}
