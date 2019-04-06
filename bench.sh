#!/usr/bin/env bash

mkdir -p perf-data

usage() {
    echo "Usage $0 [package name = point, space, all]"
    exit 1
}

pack="$1"
if [[ -z "$pack" ]]; then
    usage
fi

runAndSave() {
    local packageName="$1"
    go test -parallel 4 -cpuprofile perf-data/cpu-${packageName}.prof -memprofile perf-data/mem-${packageName}.prof -run='^$' -bench=. ./m3${packageName}/ >> ./docs/${packageName}-BenchResults.txt
    go tool pprof --text perf-data/cpu-${packageName}.prof > ./docs/${packageName}-BenchPprofCpuResults.txt
    go tool pprof --text perf-data/mem-${packageName}.prof > ./docs/${packageName}-BenchPprofMemResults.txt
}

if [ "$pack" == "point" ] || [ "$pack" == "space" ]; then
    runAndSave ${pack}
    exit $?
fi

if [ "$pack" == "all" ]; then
    runAndSave point
    runAndSave space
    exit $?
fi

echo "Package $pack unknown"
usage