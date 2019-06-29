#!/usr/bin/env bash

usage() {
    echo "Usage qsm run [filldb, gentxt, play]"
    exit 1
}

if [[ -z "$1" ]]; then
    usage
fi

if [ "$1" != "play" ] && [ "$1" == "gentxt" ] && [ "$1" != "filldb" ]; then
    echo "ERROR: Run command $1 unknown"
    usage
fi

go build && ./qsm-go $@
