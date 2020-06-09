#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="${SCRIPT_DIR}/../"

cd ${PROJECT_DIR}

echo "Running End to End tests. This requires skaffold and ko. 'kind' is required for running locally"
go build ./cmd/jobber
go test ./test/e2e/... --tags=e2e  --count=1 -v