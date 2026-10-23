#!/usr/bin/env bash
# cargo run -p qsm-view --features view -- <simulation name>
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$here/functions.sh"

if [[ $# -lt 1 ]]; then
  echo "usage: run view <simulation name> [--screenshot <png>]" >&2
  exit 2
fi

qsm_prepare
echo "cargo run -p qsm-view --features view -- $*"
echo "data: $QSM_DATA_ROOT"
exec cargo run -p qsm-view --features view -- "$@"
