#!/usr/bin/env bash

usage() {
  echo "Usage qsm run [tidy, build, filldb, gentxt, play, perf]"
  exit 1
}

if [[ -z "$1" ]]; then
  usage
fi

curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./functions.sh
. "$curDir/functions.sh"
if [[ $? -ne 0 ]]; then
  echo "ERROR: failed to load functions at $curDir/functions.sh"
  exit 2
fi

commandName=$1
shift

launch_ui() {
  local env_file="local-env.env"
  if [[ "$1" == "-kinto" ]]; then
    env_file="kinto-env.env"
    shift
  fi
  if [[ "$1" == "-okteto" ]]; then
    env_file="okteto-env.env"
    shift
  fi
  cd "${rootDir}/ui" && ${go_exe} build && cp "$env_file" ".env" && ./ui "$@"
}

case "$commandName" in
tidy)
  for m in m3util model backend client; do
    echo "INFO: Tyding go.mod of $m"
    cd ${rootDir}/${m} && ${go_exe} mod tidy
  done
  ;;
build)
  cd ${rootDir}/backend && ${go_exe} build &&
    cd ${rootDir}/ui && ${go_exe} build
  ;;
play)
  launch_ui "$@"
  ;;
gentxt | *filldb | perf)
  cd ${rootDir}/backend && ${go_exe} build && ./backend "$@"
  ;;
*)
  echo "ERROR: Run command $commandName unknown"
  usage
  ;;
esac
