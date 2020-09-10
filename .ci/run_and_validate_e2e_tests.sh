set -euo pipefail
SCRIPT_DIRPATH="$(cd "$(dirname "${0}")" && pwd)"
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"

DOCKER_REPO="avaplatform"

# login to AWS for byzantine images
echo "$DOCKER_PASS" | docker login --username "$DOCKER_USERNAME" --password-stdin

DEFAULT_CONTROLLER_TAG="$DOCKER_REPO/avalanche-testing_controller"

# Use stable version of Everest for CI
AVALANCHE_IMAGE="$DOCKER_REPO/avalanche-go:everest-latest"
# Use stable version of avalanche-byzantine based on everest for CI
BYZANTINE_IMAGE="$DOCKER_REPO/avalanche-byzantine:everest-latest"

# Kurtosis will try to pull Docker images, but as of 2020-08-09 it doesn't currently support pulling from Docker repos that require authentication
# so we have to do the pull here
docker pull "${BYZANTINE_IMAGE}"
docker pull "${AVALANCHE_IMAGE}"

E2E_TEST_COMMAND="${ROOT_DIRPATH}/scripts/full_rebuild_and_run.sh"
BYZANTINE_IMAGE_ARG="--byzantine-image-name=${BYZANTINE_IMAGE}"
AVALANCHE_IMAGE_ARG="--avalanche-image-name=${AVALANCHE_IMAGE}"
return_code=0
if ! bash "${E2E_TEST_COMMAND}" "${BYZANTINE_IMAGE_ARG}" "${AVALANCHE_IMAGE_ARG}"; then
    echo "Avalanche E2E tests failed"
    return_code=1
else
    echo "Avalanche E2E tests succeeded"
    return_code=0
fi

# Clear containers.
echo "Clearing Avalanche Docker containers..."
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${AVALANCHE_IMAGE}" --format="{{.ID}}")) >/dev/null
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${BYZANTINE_IMAGE}" --format="{{.ID}}")) >/dev/null
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${DEFAULT_CONTROLLER_TAG}" --format="{{.ID}}")) >/dev/null
echo "Avalanche Docker containers cleared successfully"

exit "${return_code}"
