#!/usr/bin/env bash

mkdir -p perf-data

pack="$1"
if [[ -z "$pack" ]]; then
    echo "Usage $0 [package name = point, space, gl]"
    exit 1
fi

runAndSave() {
    local packageName="$1"
    go test -parallel 4 -cpuprofile perf-data/cpu-${packageName}.prof -memprofile perf-data/mem-${packageName}.prof -run='^$' -bench=. ./m3${packageName}/ >> ./docs/${packageName}-BenchResults.txt
    go tool pprof --text perf-data/cpu-${packageName}.prof > ./docs/${packageName}-BenchPprofCpuResults.txt
    go tool pprof --text perf-data/mem-${packageName}.prof > ./docs/${packageName}-BenchPprofMemResults.txt
}

if [ "$pack" == "point" ] || [ "$pack" == "space" ] || [ "$pack" == "gl" ]; then
    runAndSave ${pack}
    exit $?
fi

echo "Usage $0 [package name = point, space, gl]"
echo "Package $pack unknown"
exit 2
