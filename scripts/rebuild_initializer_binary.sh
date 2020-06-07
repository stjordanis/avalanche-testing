#!/bin/bash

set -euo pipefail

SCRIPT_DIRPATH=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
ROOT_DIRPATH=$(dirname "${SCRIPT_DIRPATH}")

BUILD_DIR="build"
MAIN_BINARY_OUTPUT_FILE="ava-e2e-tests"
MAIN_BINARY_OUTPUT_PATH="${ROOT_DIRPATH}/${BUILD_DIR}/${MAIN_BINARY_OUTPUT_FILE}"

echo "Running unit tests..."
go test "${ROOT_DIRPATH}"/...
echo "Building..."
go build -o "${MAIN_BINARY_OUTPUT_PATH}" "${ROOT_DIRPATH}/initializer/main.go"
EXIT_STATUS=$?

if [ "${EXIT_STATUS}" -eq "0" ]; then
        echo "Build Successful"
        echo "Built initializer binary at ${MAIN_BINARY_OUTPUT_PATH}"
        echo "Run '${MAIN_BINARY_OUTPUT_PATH} --help' for usage."
else
        echo "Build failure"
        exit 1
fi
