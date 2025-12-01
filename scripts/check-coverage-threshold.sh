#!/usr/bin/env bash
set -euo pipefail

coverage_file=${1:-coverage/coverage.out}
threshold=${2:-30}

if [[ ! -f "${coverage_file}" ]]; then
  echo "coverage file not found: ${coverage_file}" >&2
  exit 1
fi

total_line=$(go tool cover -func="${coverage_file}" | grep '^total:')
if [[ -z "${total_line}" ]]; then
  echo "unable to parse total coverage from: ${coverage_file}" >&2
  exit 1
fi

coverage_value=$(echo "${total_line}" | awk '{print $3}')
coverage_number=${coverage_value%\%}

if [[ -z "${coverage_number}" ]]; then
  echo "parsed empty coverage value from: ${coverage_file}" >&2
  exit 1
fi

comparison=$(awk -v cov="${coverage_number}" -v thr="${threshold}" 'BEGIN { if (cov < thr) print "lt"; else print "ge"; }')

if [[ "${comparison}" == "lt" ]]; then
  echo "Coverage ${coverage_value} is below the required threshold of ${threshold}%." >&2
  exit 1
fi

echo "Coverage ${coverage_value} meets the required threshold of ${threshold}%."
