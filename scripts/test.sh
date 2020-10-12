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
curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./functions.sh
. "$curDir/functions.sh"

test_model() {
  cd ${rootDir}/model && go test ./m3point/
}

test_client() {
  cd ${rootDir}/client && go test ./clpoint/ ./clpath/ ./clspace/
}

test_backend() {
  cd ${rootDir}/backend && go test ./m3db/ ./pointdb/ ./pathdb/ ./spacedb/ ./m3server/
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
model | client | ui | backend | perf)
  test_${pack}
  ;;
all)
  ${rootDir}/qsm server stop
  test_model && test_backend && ${rootDir}/qsm server launch && test_client && test_ui
  ${rootDir}/qsm server stop
  ;;
*)
  usage
  ;;
esac
