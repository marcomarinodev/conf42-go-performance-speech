# conf42-go-performance-speech
Conf42 Golang speech - Optimizing Go Performance: Tips and Techniques


## Presentation Overview
*This repository contains materials for a conference talk on optimizing Go performance. The talk covers best practices, optimization techniques, and live coding demonstrations.*

### Key Topics Covered:
* Understanding Go Performance Factors
* Optimization Strategies for Go Applications
* Live Coding Demonstration
* Best Practices and Trade-offs

## Execute the benchmarks
To execute the benchmarks, **go** in the test folder and run:

```bash
chmod 777 gen-test.sh
./gen-test.sh 
```

The output will be put in:
- *benchstat* folder where you can find the benchstat comparison between the different pipelines
- *profiles* folder where there are the cpu and mem profiles

then you can compare the results using *benchstat*

```bash
cd benchstat
benchstat seq.txt first_opt.txt sec_opt.txt
```

you're going to get a result like the following:
```log
goos: darwin
goarch: arm64
pkg: transactions_sample/test
BenchmarkCustomPipeline_Seq-10                 	      30	  46483878 ns/op	311761696 B/op	      39 allocs/op
BenchmarkCustomPipeline_FirstOpt-10            	     100	  13113456 ns/op	234776845 B/op	     640 allocs/op
BenchmarkCustomPipeline_Optimized_SecOpt-10    	   12980	     79316 ns/op	  120215 B/op	     328 allocs/op
PASS
ok  	transactions_sample/test	49.250s

```

### Using pprof to read the profiles
Let's say we want to read the cpu profile of `cpu_seq.prof` the go tool command for pprof is:
```bash
go tool pprof cpu_seq.prof
```

then follow the commands provided in the official [pprof](https://github.com/google/pprof) docs.


