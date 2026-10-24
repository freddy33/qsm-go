#!/usr/bin/env bash
# cargo run -p qsm -- <simulation name> [--new]
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$here/functions.sh"

if [[ $# -lt 1 ]]; then
  echo "usage: run run <simulation name> [--new]" >&2
  exit 2
fi

qsm_prepare
echo "cargo run -p qsm -- $*"
echo "data: $QSM_DATA_ROOT"
exec cargo run -p qsm -- "$@"
