#!/bin/bash

set -euo pipefail
SCRIPT_DIRPATH=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

# TODO this should be avalabs!
DOCKER_ORG="kurtosistech"
REPO_BASE="ava-e2e-tests"
CONTROLLER_REPO="${REPO_BASE}_controller"
LATEST_TAG="latest"

ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"
DOCKER="${DOCKER:-docker}"

COMMIT="$(git --git-dir="${ROOT_DIRPATH}/.git" rev-parse --short HEAD)"

LATEST_CONTROLLER_TAG="${DOCKER_ORG}/${CONTROLLER_REPO}:${LATEST_TAG}"
"${DOCKER}" build -t "${LATEST_CONTROLLER_TAG}" "${ROOT_DIRPATH}" -f "${ROOT_DIRPATH}/controller/Dockerfile"
