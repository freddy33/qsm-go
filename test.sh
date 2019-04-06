#!/usr/bin/env bash

usage() {
    echo "Usage $0 [package name = point, space, gl, all]"
    exit 1
}

pack="$1"
if [[ -z "$pack" ]]; then
    usage
fi

runTest() {
    local packageName="$1"
    go test -parallel 4 ./m3${packageName}/
}

if [ "$pack" == "point" ] || [ "$pack" == "space" ] || [ "$pack" == "gl" ]; then
    runTest ${pack}
    exit $?
fi

if [ "$pack" == "all" ]; then
    runTest point
    runTest space
    runTest gl
    exit $?
fi

echo "Package $pack unknown"
usage