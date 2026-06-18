#!/usr/bin/env bash
# Run port contract tests (no Docker) across data-plane modules.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

run_contract() {
  local module="$1"
  local path="${ROOT}/${module}"
  if [[ ! -d "${path}/contract" ]]; then
    return 0
  fi
  echo "=== contract ${module} ==="
  (cd "${path}" && go test ./contract/... -race -count=1)
}

for module in cache database mq secret storage task; do
  run_contract "${module}"
done
