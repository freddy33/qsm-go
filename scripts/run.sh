#!/usr/bin/env bash

usage() {
    echo "Usage qsm run [refilldb, filldb, gentxt, play]"
    exit 1
}

if [[ -z "$1" ]]; then
    usage
fi

if [ "$1" != "play" ] && [ "$1" == "gentxt" ] && [ "$1" != "filldb" ] && [ "$1" != "refilldb" ]; then
    echo "ERROR: Run command $1 unknown"
    usage
fi

curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
# shellcheck source=./functions.sh
. "$curDir/functions.sh"

$go_exe build && ./qsm-go$exe_ext $@
