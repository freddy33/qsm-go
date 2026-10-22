#!/usr/bin/env bash
# cargo test --workspace, from any working directory.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$here/functions.sh"
qsm_prepare

echo "cargo test --workspace"
echo "data: $QSM_DATA_ROOT"
exec cargo test --workspace "$@"
