#!/bin/bash

set -euo pipefail
SCRIPT_DIRPATH=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"
DOCKER="${DOCKER:-docker}"

if [ ${#} -eq 0 ]; then
    COMMIT="$(git --git-dir="${ROOT_DIRPATH}/.git" rev-parse --short HEAD)"
else
    COMMIT="${1}"
fi

TAG="ava-test-controller:${COMMIT}"
"${DOCKER}" build -t "${TAG}" "${ROOT_DIRPATH}" -f "${ROOT_DIRPATH}/controller/Dockerfile"
