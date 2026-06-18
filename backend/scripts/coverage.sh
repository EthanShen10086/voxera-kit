#!/usr/bin/env bash
# Merge per-module Go coverage profiles from go.work modules.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${ROOT}/coverage"
mkdir -p "${OUT}"

merged="${OUT}/merged.out"
echo "mode: atomic" > "${merged}"

while IFS= read -r dir; do
  module="${dir#./}"
  echo "=== coverage ${module} ==="
  profile="${OUT}/${module//\//-}.out"
  if (cd "${ROOT}/${module}" && go test ./... -coverprofile="${profile}" -covermode=atomic); then
    if [[ -s "${profile}" ]]; then
      tail -n +2 "${profile}" >> "${merged}"
    fi
  else
    echo "warning: coverage failed for ${module}" >&2
    exit 1
  fi
done < <(awk '/\.\//{gsub(/[[:space:]]/, ""); gsub(/\)/, ""); if(/^\.\//) print}' "${ROOT}/go.work")

go tool cover -func="${merged}" | tail -1
echo "merged profile: ${merged}"
