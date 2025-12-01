#!/usr/bin/env bash
set -euo pipefail

coverage_file=${1:-coverage/coverage.out}

if [[ ! -f "${coverage_file}" ]]; then
  echo "coverage file not found: ${coverage_file}" >&2
  exit 1
fi

total_line=$(go tool cover -func="${coverage_file}" | grep '^total:')
coverage_value=$(echo "${total_line}" | awk '{print $3}')

echo "${coverage_value}"
