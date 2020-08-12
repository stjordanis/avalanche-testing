#!/bin/bash

set -euo pipefail

# Note: this script will build a docker image by cloning a remote version of avalanche-e2e-tests and gecko into a temporary
# location and using that version's Dockerfile to build the image.
SCRIPT_DIRPATH=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"
E2E_COMMIT="$(git --git-dir="$ROOT_DIRPATH/.git" rev-parse --short HEAD)"

export GOPATH="$SCRIPT_DIRPATH/.build_image_gopath"
WORKPREFIX="$GOPATH/src/github.com/ava-labs"
DOCKER="${DOCKER:-docker}"


# TODO set this to use the public repo and master branch, leave for now to avoid passing args while using
GECKO_REMOTE="https://github.com/ava-labs/gecko-internal.git"
GECKO_BRANCH="lock-everest"

# Clones the remote and checks out the current local commit
# Note: commit and push changes before running this script, or some local changes
# will be left out
E2E_REMOTE="https://github.com/ava-labs/avalanche-e2e-tests.git"

# Clone the remotes and checkout the desired branch/commits
GECKO_CLONE="$WORKPREFIX/gecko"
E2E_CLONE="$WORKPREFIX/avalanche-e2e-tests"

# Create the WORKPREFIX directory if it does not exist yet
if [[ ! -d "$WORKPREFIX" ]]; then
    mkdir -p "$WORKPREFIX"
fi

# Configure git credential helper
git config --global credential.helper cache

if [[ ! -d "$GECKO_CLONE" ]]; then
    git clone "$GECKO_REMOTE" "$GECKO_CLONE"
fi

git -C "$GECKO_CLONE" checkout "$GECKO_BRANCH"

GECKO_COMMIT="$(git -C "$GECKO_CLONE" rev-parse --short HEAD)"

if [[ ! -d "$E2E_CLONE" ]]; then
    git clone "$E2E_REMOTE" "$E2E_CLONE"
fi

git -C "$E2E_CLONE" checkout "$E2E_COMMIT"


DOCKER_ORG="avaplatform"
REPO_BASE="avalanche-e2e-tests"
CONTROLLER_REPO="${REPO_BASE}_controller"

CONTROLLER_TAG="$DOCKER_ORG/$CONTROLLER_REPO-$E2E_COMMIT-$GECKO_COMMIT"

"${DOCKER}" build -t "${CONTROLLER_TAG}" "${WORKPREFIX}" -f "$ROOT_DIRPATH/controller/local.Dockerfile"
