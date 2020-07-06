#!/usr/bin/env bash

usage() {
    echo "Usage qsm run [build, filldb, gentxt, play, perf]"
    exit 1
}

if [[ -z "$1" ]]; then
    usage
fi

curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
# shellcheck source=./functions.sh
. "$curDir/functions.sh"
if [[ $? -ne 0 ]]; then
    echo "ERROR: failed to load functions at $curDir/functions.sh"
    exit 2
fi

commandName=$1

case "$commandName" in
    build)
    cd ${rootDir}/backend && ${go_exe} build && \
    cd ${rootDir}/ui && ${go_exe} build
    ;;
    play)
    cd ${rootDir}/ui && ${go_exe} build && ./ui $@
    ;;
    gentxt|*filldb|perf)
    cd ${rootDir}/backend && ${go_exe} build && ./backend $@
    ;;
    *)
    echo "ERROR: Run command $commandName unknown"
    usage
    ;;
esac