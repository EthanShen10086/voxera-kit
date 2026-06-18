#!/usr/bin/env bash
# Merge per-module Go coverage profiles from go.work modules.
#
# Environment:
#   MIN_COVERAGE     — fail if merged total is below this % (default: 50, ramp见 docs/COVERAGE_ROADMAP.md)
#   COVERAGE_SKIP    — space-separated module names to skip (default: testkit)
#   COVERAGE_ENFORCE — "true" (default) to exit non-zero when below MIN_COVERAGE
#   COVERAGE_TAGS    — optional go test -tags value (e.g. integration for testcontainers adapters)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${ROOT}/coverage"
mkdir -p "${OUT}"

MIN_COVERAGE="${MIN_COVERAGE:-50}"
COVERAGE_ENFORCE="${COVERAGE_ENFORCE:-true}"
COVERAGE_SKIP="${COVERAGE_SKIP:-testkit}"
COVERAGE_TAGS="${COVERAGE_TAGS:-}"

GO_TEST_FLAGS=(-covermode=atomic)
if [[ -n "${COVERAGE_TAGS}" ]]; then
  GO_TEST_FLAGS=(-tags="${COVERAGE_TAGS}" "${GO_TEST_FLAGS[@]}")
  echo "coverage: using go test -tags=${COVERAGE_TAGS}"
fi

should_skip() {
  local module="$1"
  for skip in ${COVERAGE_SKIP}; do
    if [[ "${module}" == "${skip}" ]]; then
      return 0
    fi
  done
  return 1
}

merged="${OUT}/merged.out"
echo "mode: atomic" > "${merged}"

while IFS= read -r dir; do
  module="${dir#./}"
  if should_skip "${module}"; then
    echo "=== skip coverage ${module} (COVERAGE_SKIP) ==="
    continue
  fi
  echo "=== coverage ${module} ==="
  profile="${OUT}/${module//\//-}.out"
  if (cd "${ROOT}/${module}" && go test ./... "${GO_TEST_FLAGS[@]}" -coverprofile="${profile}"); then
    if [[ -s "${profile}" ]]; then
      tail -n +2 "${profile}" >> "${merged}"
    fi
  else
    echo "warning: coverage failed for ${module}" >&2
    exit 1
  fi
done < <(awk '/\.\//{gsub(/[[:space:]]/, ""); gsub(/\)/, ""); if(/^\.\//) print}' "${ROOT}/go.work")

summary="$(go tool cover -func="${merged}" | tail -1)"
echo "${summary}"
echo "merged profile: ${merged}"

pct="$(echo "${summary}" | awk '{print $NF}' | tr -d '%')"
if [[ -z "${pct}" ]]; then
  echo "error: could not parse coverage percentage" >&2
  exit 1
fi

milestones=(15 30 50 80)
for target in "${milestones[@]}"; do
  if awk "BEGIN {exit !(${pct} < ${target})}"; then
    echo "coverage ramp: next milestone ${target}% (see docs/COVERAGE_ROADMAP.md)"
    break
  fi
done

if awk "BEGIN {exit !(${pct} < ${MIN_COVERAGE})}"; then
  msg="merged coverage ${pct}% is below MIN_COVERAGE=${MIN_COVERAGE}%"
  if [[ "${COVERAGE_ENFORCE}" == "true" ]]; then
    echo "error: ${msg}" >&2
    exit 1
  fi
  echo "warning: ${msg}" >&2
else
  echo "coverage gate: ${pct}% >= MIN_COVERAGE=${MIN_COVERAGE}%"
fi
