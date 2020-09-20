#!/bin/bash

set -euo pipefail

# Note: this script will build a docker image by cloning a remote version of avalanche-testing and avalanchego into a temporary
# location and using that version's Dockerfile to build the image.
SCRIPT_DIRPATH=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"
AVALANCHE_PATH="$GOPATH/src/github.com/ava-labs/avalanchego"
E2E_COMMIT="$(git --git-dir="$ROOT_DIRPATH/.git" rev-parse --short HEAD)"
AVALANCHE_COMMIT="$(git --git-dir="$AVALANCHE_PATH/.git" rev-parse --short HEAD)"

export GOPATH="$SCRIPT_DIRPATH/.build_image_gopath"
WORKPREFIX="$GOPATH/src/github.com/ava-labs"
DOCKER="${DOCKER:-docker}"


AVALANCHE_REMOTE="https://github.com/ava-labs/avalanchego.git"
E2E_REMOTE="https://github.com/ava-labs/avalanche-testing.git"


# Clone the remotes and checkout the desired branch/commits
AVALANCHE_CLONE="$WORKPREFIX/avalanchego"
E2E_CLONE="$WORKPREFIX/avalanche-testing"

# Create the WORKPREFIX directory if it does not exist yet
if [[ ! -d "$WORKPREFIX" ]]; then
    mkdir -p "$WORKPREFIX"
fi

# Configure git credential helper
git config --global credential.helper cache

if [[ ! -d "$AVALANCHE_CLONE" ]]; then
    git clone "$AVALANCHE_REMOTE" "$AVALANCHE_CLONE"
else
    git -C "$AVALANCHE_CLONE" fetch origin
fi

git -C "$AVALANCHE_CLONE" checkout "$AVALANCHE_COMMIT"

if [[ ! -d "$E2E_CLONE" ]]; then
    git clone "$E2E_REMOTE" "$E2E_CLONE"
else
    git -C "$E2E_CLONE" fetch origin
fi

git -C "$E2E_CLONE" checkout "$E2E_COMMIT"


DOCKER_ORG="avaplatform"
REPO_BASE="avalanche-testing"
CONTROLLER_REPO="${REPO_BASE}_controller"

CONTROLLER_TAG="$DOCKER_ORG/$CONTROLLER_REPO-$E2E_COMMIT-$AVALANCHE_COMMIT"

"${DOCKER}" build -t "${CONTROLLER_TAG}" "${WORKPREFIX}" -f "$ROOT_DIRPATH/testsuite/local.Dockerfile"
