#!/usr/bin/env bash

usage() {
    echo "Usage qsm test [package name = point, path, space, gl, db, all, perf]"
    exit 1
}

pack="$1"
if [[ -z "$pack" ]]; then
    usage
fi

curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
# shellcheck source=./functions.sh
. "$curDir/functions.sh"

if [ "$pack" == "point" ] || [ "$pack" == "path" ] || [ "$pack" == "space" ] || [ "$pack" == "db" ] || [ "$pack" == "gl" ]; then
    $go_exe test ./m3${pack}/
    exit $?
fi

if [ "$pack" == "all" ]; then
    $go_exe test -parallel 4 ./m3db/ ./m3point/ ./m3path/ ./m3space/ ./m3gl/
    exit $?
fi

if [ "$pack" == "perf" ]; then
    # Performance test is 3
    export QSM_ENV_NUMBER=3

    dbLoc="was-not-set"
    confDir="was-not-set"
    . ./scripts/functions.sh

    ./qsm db stop
    cp $confDir/postgresql.conf $dbLoc/postgresql.conf && ./qsm db drop && ./qsm run filldb
    if [ $? -ne 0 ]; then
        echo "ERROR: Setting perf DB failed!"
        exit 13
    fi
    export GOMAXPROCS=50
    ./qsm run perf
    if [ $? -ne 0 ]; then
        echo "ERROR: executing perf DB test returned error"
        exit 3
    fi
    exit 0
fi

echo "Package $pack unknown"
usage
