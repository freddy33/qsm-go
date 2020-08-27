#!/usr/bin/env bash

usage() {
  echo "Usage qsm server [term, launch, test, stop, stoptest]"
  exit 1
}

if [[ -z "$1" ]]; then
  usage
fi

go_exe=go
rootDir="was-not-set"
logDir="was-not-set"
curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./functions.sh
. "$curDir/functions.sh"
if [[ $? -ne 0 ]]; then
  echo "ERROR: failed to load functions at $curDir/functions.sh"
  exit 2
fi

port="8063"
isTest=false

analyzeParameters() {
  local portParam
  port="8063"
  for arg in "$@"; do
    if [[ "$arg" == "-test" ]]; then
      isTest=true
      port="8877"
      break
    fi
  done
  portParam=false
  for arg in "$@"; do
    if ${portParam}; then
      port=arg
      break
    fi
    if [[ "$arg" == "-port" ]]; then
      portParam=true
    fi
  done
}

getPidFile() {
  if ${isTest}; then
    echo "${logDir}/backend-test-${port}.pid"
  else
    echo "${logDir}/backend-${port}.pid"
  fi
}

getLogFile() {
  local outLog
  if ${isTest}; then
    outLog="${logDir}/backend-test-${port}.log"
  else
    outLog="${logDir}/backend-${port}.log"
  fi
  if [[ -e "$outLog" ]]; then
    mv "$outLog" "${outLog:0:-4}-$(date "+%Y-%m-%d_%H-%M-%S").log"
  fi
  echo "${outLog}"
}

stopServer() {
  local pidFile
  local pidValue
  analyzeParameters "$@"
  pidFile="$(getPidFile)"
  if [[ -e "$pidFile" ]]; then
    pidValue=$(cat "${pidFile}")
    echo "Stopping server at pid $pidValue for port $port"
    # shellcheck disable=SC2086
    kill ${pidValue}
    sleep 1
    rm "${pidFile}"
  fi
}

runServer() {
  local logFile
  local backendPid
  analyzeParameters "$@"
  pidFile="$(getPidFile)"
  if [[ -e "$pidFile" ]]; then
    echo "Server on  port $port already running"
    exit 2
  fi
  logFile="$(getLogFile)"
  echo "INFO: Launching backend QSM log file at=$logFile"
  cd "${rootDir}/backend" && ${go_exe} build || exit $?
  nohup ./backend server "$@" 1>"${logFile}" 2>&1 &
  backendPid="$!"
  echo "INFO: Backend QSM launched PID=$backendPid"
  echo "$backendPid" >"$pidFile"
  sleep 1
}

runBackend() {
  cd "${rootDir}/backend" && ${go_exe} build && ./backend server "$@"
}

commandName=$1
shift

case "$commandName" in
term)
  runBackend "$@"
  ;;
launch)
  runServer "$@"
  ;;
test)
  runServer -test "$@"
  ;;
stop)
  stopServer "$@"
  ;;
stoptest)
  stopServer -test "$@"
  ;;
*)
  usage
  ;;
esac
