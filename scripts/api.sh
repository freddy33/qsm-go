#!/usr/bin/env bash

perfDir="build/perf-data"
logDir="build/log"
mkdir -p $perfDir
mkdir -p $logDir

usage() {
  echo "Usage qsm api"
  exit 1
}

dt=$(date '+%Y%m%d_%H%M%S')
echo "$dt"

command="$1"
shift

baseUrl="http://localhost:3002"

if [[ "$1" == "-kinto" ]]; then
  baseUrl="https://qsmgo-92a1656-5f154.eu1.kinto.io"
  shift
fi
if [[ "$1" == "-okteto" ]]; then
  baseUrl="https://backend-freddy33.cloud.okteto.net"
  shift
fi

QSM_ENV_NUMBER=${QSM_ENV_NUMBER:=1}

baseCurlCommand="curl -s -f -H \"QsmEnvId: $QSM_ENV_NUMBER\""
baseJsonCurlCommand="$baseCurlCommand -H \"Accept: application/json\""

drop_env() {
  if [[ "$1" != "qsm$QSM_ENV_NUMBER" ]]; then
    echo "ERROR: For delete please give full schema name"
    exit 2
  fi
  ${baseCurlCommand} -X DELETE "$baseUrl/drop-env"
}

list_env() {
  ${baseJsonCurlCommand} "$baseUrl/list-env" | jq ".$1"
}

init_env() {
  ${baseCurlCommand} -X POST "$baseUrl/init-env"
}

list_path_context() {
  local ctx_id="-1"
  if [[ -n "$1" ]]; then
    ctx_id="$1"
  fi
  ${baseJsonCurlCommand} "$baseUrl/path-context?path_ctx_id=$ctx_id" | jq ".$2"
}

increase_path_context() {
  if [[ -z "$1" ]] || [[ -z "$2" ]]; then
    echo "ERROR: Need a path context id param to increase max dist"
  fi
  ${baseJsonCurlCommand} -X PUT "$baseUrl/max-dist?path_ctx_id=$1&dist=$2" | jq ".$3"
}

increase_path_context_to() {
  if [[ -z "$1" ]] || [[ -z "$2" ]]; then
    echo "ERROR: Need a path context id param to increase max dist"
  fi
  local current_dist=$(list_path_context $1 max_dist)
  if [[ $? -ne 0 ]]; then
    echo "ERROR: Received error retrieving current max dist"
    exit 3
  fi
  echo "INFO: Received current max dist $current_dist for path $1"
  ((start = $current_dist + 1))
  TIMEFORMAT=%R
  for dist in $(seq $start $2); do
    echo "INFO: Requesting for path_ctx_id=$1 the dist=$dist"
    time ${baseJsonCurlCommand} -X PUT "$baseUrl/max-dist?path_ctx_id=$1&dist=$dist" | jq ".nb_path_nodes"
    if [[ $? -ne 0 ]]; then
      echo "ERROR: Received error increasing current max dist"
      exit 4
    fi
  done
}

case "$command" in
list)
  list_env "$@"
  ;;
init)
  init_env "$@"
  ;;
drop)
  drop_env "$@"
  ;;
path)
  list_path_context "$@"
  ;;
increase)
  increase_path_context "$@"
  ;;
inc_to)
  increase_path_context_to "$@"
  ;;
*)
  usage
  ;;
esac
