#!/usr/bin/env bash

usage() {
    echo "Usage qsm test [package name = point, path, space, gl, db, all]"
    exit 1
}

pack="$1"
if [[ -z "$pack" ]]; then
    usage
fi

# TODO: analyze second param to test on other env. By default it's 3
export QSM_ENV_NUMBER=3

if [ "$pack" == "point" ] || [ "$pack" == "path" ] || [ "$pack" == "space" ] || [ "$pack" == "db" ] || [ "$pack" == "gl" ]; then
    go test ./m3${pack}/
    exit $?
fi

if [ "$pack" == "all" ]; then
    go test -parallel 4 ./m3point/ ./m3path/ ./m3db/ ./m3space/ ./m3gl/
    exit $?
fi

echo "Package $pack unknown"
usage