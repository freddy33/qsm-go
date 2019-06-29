#!/usr/bin/env bash

perfDir="build/perf-data"
logDir="build/log"
mkdir -p $perfDir
mkdir -p $logDir

usage() {
    echo "Usage qsm bench [package name = path, space, all]"
    exit 1
}

pack="$1"
if [[ -z "$pack" ]]; then
    usage
fi

dt=$(date '+%Y%m%d_%H%M%S');
echo "$dt"

runAndSave() {
    local packageName="$1"
    go test -parallel 4 -cpuprofile $perfDir/cpu-${packageName}.prof -memprofile $perfDir/mem-${packageName}.prof -run='^$' -bench=. ./m3${packageName}/ >> ./docs/${packageName}-BenchResults.txt
    go tool pprof --text $perfDir/cpu-${packageName}.prof > ./docs/${packageName}-BenchPprofCpuResults.txt
    go tool pprof --text $perfDir/mem-${packageName}.prof > ./docs/${packageName}-BenchPprofMemResults.txt
}

runSimple() {
    local packageName="$1"
    go test -parallel 4 -run='^$' -bench=. ./m3${packageName}/ >> $logDir/${packageName}-Bench-$dt.log
}

if [ "$pack" == "path" ] || [ "$pack" == "space" ]; then
    if [ "$2" == "-s" ]; then
        runSimple ${pack}
    else
        runAndSave ${pack}
    fi
    exit $?
fi

if [ "$pack" == "all" ]; then
    runAndSave path
    runAndSave space
    exit $?
fi

echo "Package $pack unknown"
usage