#!/usr/bin/env bash

usage() {
    echo "Usage qsm test [package name = util, model, ui, backend, all, perf]"
    exit 1
}

pack="$1"
if [[ -z "$pack" ]]; then
    usage
fi
shift

dbLoc="was-not-set"
confDir="was-not-set"
curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
# shellcheck source=./functions.sh
. "$curDir/functions.sh"

test_util() {
    cd ${rootDir}/utils && go test ./m3db/
}

test_model() {
    cd ${rootDir}/model && go test ./m3point/ ./m3path/ ./m3space/
}

test_backend() {
    cd ${rootDir}/backend && go test ./m3api/
}

test_ui() {
    cd ${rootDir}/ui && go test ./m3gl/
}

test_perf() {
    # Performance test is 3
    export QSM_ENV_NUMBER=3

    ${rootDir}/qsm db stop
    cp $confDir/postgresql.conf $dbLoc/postgresql.conf && ./qsm db drop && ./qsm run filldb
    if [ $? -ne 0 ]; then
        echo "ERROR: Setting perf DB failed!"
        return 13
    fi
    export GOMAXPROCS=50
    ${rootDir}/qsm run perf
    if [ $? -ne 0 ]; then
        echo "ERROR: executing perf DB test returned error"
        return 3
    fi
    return 0
}

case "$pack" in
    util|model|ui|backend|perf)
    test_${pack}
    ;;
    all)
    #test_util && test_model && test_backend && test_ui
    test_util && test_model && test_ui
    ;;
    *)
    usage
    ;;
esac
