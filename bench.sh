#!/usr/bin/env bash

go test -parallel 4 -cpuprofile cpu.prof -memprofile mem.prof -run='^$' -bench=. ./m3space/ >> ./docs/BenchResults.txt
go tool pprof --text cpu.prof > ./docs/BenchPprofCpuResults.txt
go tool pprof --text mem.prof > ./docs/BenchPprofMemResults.txt
