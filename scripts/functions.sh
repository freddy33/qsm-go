#!/usr/bin/env bash
# Shared paths for the QSM scripts. Source this file; do not execute it.

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  echo "source scripts/functions.sh from the other scripts" >&2
  exit 2
fi

# Repo root. This file lives in scripts/.
qsm_root() {
  local here
  here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  (cd "$here/.." && pwd)
}

# Move to the Cargo workspace and pin the data directory.
# An unset QSM_DATA_ROOT becomes <repo>/data. A relative value is resolved
# from the caller's working directory before the cd.
qsm_prepare() {
  local root caller
  caller="$(pwd)"
  root="$(qsm_root)"
  export QSM_HOME="$root"
  if [[ -z "${QSM_DATA_ROOT:-}" ]]; then
    export QSM_DATA_ROOT="$root/data"
  elif [[ "$QSM_DATA_ROOT" != /* ]]; then
    QSM_DATA_ROOT="$caller/$QSM_DATA_ROOT"
    export QSM_DATA_ROOT
  fi
  cd "$root/rust" || exit 1
}
