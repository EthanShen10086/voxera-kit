#!/usr/bin/env bash
# Monthly (or kit minor) coordinated pin bump across consumers.
set -euo pipefail

ORG="${GITHUB_ORG:-EthanShen10086}"
KIT_REPO="${KIT_REPO:-${ORG}/voxera-kit}"
TAG="${1:-}"

if [[ -z "${TAG}" ]]; then
  echo "Usage: $0 <kit-tag>   e.g. $0 v0.3.0"
  exit 1
fi

PIN="$(gh api "repos/${KIT_REPO}/git/ref/tags/${TAG}" --jq '.object.sha' 2>/dev/null || true)"
if [[ -z "${PIN}" || "${PIN}" == "null" ]]; then
  # annotated tag
  PIN="$(gh api "repos/${KIT_REPO}/git/refs/tags/${TAG}" --jq '.object.sha')"
  OBJ_TYPE="$(gh api "repos/${KIT_REPO}/git/tags/${PIN}" --jq '.object.type')"
  if [[ "${OBJ_TYPE}" == "tag" ]]; then
    PIN="$(gh api "repos/${KIT_REPO}/git/tags/${PIN}" --jq '.object.sha')"
  fi
fi

echo "Kit tag ${TAG} → pin ${PIN}"

CONSUMERS=(voxera finera pulsera MsgGuard ciphera ciphera-vps)
for repo in "${CONSUMERS[@]}"; do
  path="/Users/ethanshen/Desktop/code/${repo}/.github/voxera-kit-pin"
  if [[ ! -f "${path}" ]]; then
    echo "skip ${repo} (no pin file)"
    continue
  fi
  echo "${PIN}" > "${path}"
  echo "updated ${repo}"
done

cat <<EOF

Next steps (manual):
1. In each consumer: git commit -m "chore: bump voxera-kit pin to ${TAG}"
2. Update voxera-kit docs/COMPATIBILITY.md
3. Push consumers after kit CI green
See docs/RELEASE_TRAIN.md
EOF
