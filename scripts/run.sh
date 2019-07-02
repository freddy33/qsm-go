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

go build && ./qsm-go $@
