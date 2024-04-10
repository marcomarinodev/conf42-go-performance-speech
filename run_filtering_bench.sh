#!/bin/bash

go test -bench=BenchmarkFiltering_Linear
go test -bench=BenchmarkFiltering_PrefixTree
