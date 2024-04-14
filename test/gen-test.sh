#!/bin/bash

go test -bench=BenchmarkCustomPipeline_Seq \
    -run='^$' \
    -count=6 \
    -benchtime=5s \
    -benchmem \
    -cpuprofile profiles/cpu_seq.prof \
    -memprofile profiles/mem_seq.prof \
    > benchstat/seq.txt

go test -bench=BenchmarkCustomPipeline_FirstOpt \
    -run='^$' \
    -count=6 \
    -benchtime=5s \
    -benchmem \
    -cpuprofile profiles/cpu_first_opt.prof \
    -memprofile profiles/mem_first_opt.prof \
    > benchstat/first_opt.txt

go test -bench=BenchmarkCustomPipeline_Optimized_SecOpt \
    -run='^$' \
    -count=6 \
    -benchtime=5s \
    -benchmem \
    -cpuprofile profiles/cpu_sec_opt.prof \
    -memprofile profiles/mem_sec_opt.prof \
    > benchstat/sec_opt.txt