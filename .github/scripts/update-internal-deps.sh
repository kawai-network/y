#!/usr/bin/env bash
set -euo pipefail

INTERNAL_REF="${INTERNAL_REF:-main}"
INTERNAL_PATTERNS='^github\.com/(kawai-network|getkawai)/'

module_path="$(go list -m -f '{{.Path}}')"

echo "Current module: ${module_path}"
echo "Target ref: ${INTERNAL_REF}"

deps="$({ go list -m all 2>/dev/null || true; } | awk '{print $1}' | grep -E "${INTERNAL_PATTERNS}" | grep -v "^${module_path}$" | sort -u || true)"

if [[ -z "${deps}" ]]; then
  echo "No internal dependencies found."
  exit 0
fi

echo "Internal dependencies to update:"
echo "${deps}"

while IFS= read -r dep; do
  [[ -z "${dep}" ]] && continue
  echo "Updating ${dep}@${INTERNAL_REF}"
  go get "${dep}@${INTERNAL_REF}"
done <<< "${deps}"

go mod tidy
