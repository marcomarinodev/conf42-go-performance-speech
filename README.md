# conf42-go-performance-speech
Conf42 Golang speech - Optimizing Go Performance: Tips and Techniques


## Presentation Overview

This repository contains materials for a conference talk on optimizing Go performance. The talk covers best practices, optimization techniques, and live coding demonstrations.

## Execute the benchmarks and tests
Go in the test folder and run:

```bash
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

```


To execute the benchmarks, move in the project directory and run `run_pipeline_bench.sh`.
The output will be put in:
- *bench* folder where you can find the benchstat comparison between the different pipelines
- *profiles* folder where there are the cpu and mem profiles

Or, if you prefer, you can run the following:

```bash
go test -bench=<benchmark_prefix_name> -run='^$' -count=<n_of_iterations> -benchmem -cpuprofile=<path> -memprofile=<path>
```

### Using pprof to read the profiles
Let's say we want to read the cpu profile of `cpu_seq.prof` the go tool command for pprof is:
```bash
go tool pprof cpu_seq.prof
```

### Key Topics Covered:

* Understanding Go Performance Factors
* Optimization Strategies for Go Applications
* Live Coding Demonstration
* Best Practices and Trade-offs
