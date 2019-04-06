#!/usr/bin/env bash

usage() {
    echo "Usage $0 [package name = point, space, gl, all]"
    exit 1
}

pack="$1"
if [[ -z "$pack" ]]; then
    usage
fi

if [ "$pack" == "point" ] || [ "$pack" == "space" ] || [ "$pack" == "gl" ]; then
    go test ./m3${pack}/
    exit $?
fi

if [ "$pack" == "all" ]; then
    go test -parallel 4 ./m3point/ ./m3space/ ./m3gl/
    exit $?
fi

echo "Package $pack unknown"
usage