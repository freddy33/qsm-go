#!/usr/bin/env bash

usage() {
  echo "Usage qsm server [term, launch, stop]"
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

backendDir="${rootDir}/backend"

getServerPort() {
  local envFile
  local port
  envFile="${backendDir}/.env"
  if [[ -e "$envFile" ]]; then
    port=$(grep "SERVER_PORT" "$envFile" | awk -F '=' '{print $2}')
    if [[ $? -ne 0 ]]; then
      echo "ERROR: no SERVER_PORT found in $envFile"
      exit 11
    fi
    echo "$port"
  else
    echo "ERROR: The file $envFile does not exists"
    exit 22
  fi
}

getPidFile() {
  local port
  port="$(getServerPort)"
  if [[ $? -ne 0 ]]; then
    echo "$port"
    exit 14
  fi
  echo "${logDir}/backend-${port}.pid"
}

getLogFile() {
  local port
  port="$(getServerPort)"
  if [[ $? -ne 0 ]]; then
    echo "$port"
    exit 13
  fi

  local outLog
  outLog="${logDir}/backend-${port}.log"
  if [[ -e "$outLog" ]]; then
    mv "$outLog" "${outLog:0:-4}-$(date "+%Y-%m-%d_%H-%M-%S").log"
  fi
  echo "${outLog}"
}

stopServer() {
  local port
  port="$(getServerPort)"
  if [[ $? -ne 0 ]]; then
    echo "$port"
    exit 13
  fi

  local pidFile
  local pidValue
  pidFile="$(getPidFile)"
  if [[ $? -ne 0 ]]; then
    echo "$pidFile"
    exit 15
  fi

  if [[ -e "$pidFile" ]]; then
    pidValue=$(cat "${pidFile}")
    echo "INFO: Stopping server at pid $pidValue for port $port"
    # shellcheck disable=SC2086
    kill ${pidValue}
    sleep 1
    rm "${pidFile}"
  else
    echo "DEBUG: Server for port $port already stopped"
  fi
}

runServer() {
  local pidFile
  local logFile
  local backendPid

  pidFile="$(getPidFile)"
  if [[ $? -ne 0 ]]; then
    echo "$pidFile"
    exit 16
  fi
  if [[ -e "$pidFile" ]]; then
    echo "Server on  port $port already running"
    exit 2
  fi

  logFile="$(getLogFile)"
  if [[ $? -ne 0 ]]; then
    echo "$pidFile"
    exit 16
  fi

  echo "INFO: Launching backend QSM log file at=$logFile"
  cd "${backendDir}" && ${go_exe} build || exit $?
  nohup ./backend server "$@" 1>"${logFile}" 2>&1 &
  backendPid="$!"
  echo "INFO: Backend QSM launched PID=$backendPid"
  echo "$backendPid" >"$pidFile"
  sleep 1
}

runBackend() {
  cd "${backendDir}" && ${go_exe} build && ./backend server "$@"
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
stop)
  stopServer "$@"
  ;;
*)
  usage
  ;;
esac
