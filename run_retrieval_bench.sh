#!/bin/bash

go test -bench=BenchmarkRequestSliceHandler_WithoutCache -benchmem
go test -bench=BenchmarkRequestSliceHandler_WithCache -benchmem
