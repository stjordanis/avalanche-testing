set -euo pipefail
SCRIPT_DIRPATH="$(cd "$(dirname "${0}")" && pwd)"
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"
PARALLELISM=4

DOCKER_REPO="avaplatform"

# Use stable version of Everest for CI
AVALANCHE_IMAGE="$DOCKER_REPO/avalanchego:v1.0.1"
# Use stable version of avalanche-byzantine based on everest for CI
BYZANTINE_IMAGE="$DOCKER_REPO/avalanche-byzantine:v0.1.1"

# Kurtosis will try to pull Docker images, but as of 2020-08-09 it doesn't currently support pulling from Docker repos that require authentication
# so we have to do the pull here
docker pull "${AVALANCHE_IMAGE}"

# If Docker Credentials are not available skip the Byzantine Tests
if [[ ${#DOCKER_USERNAME} == 0 ]]; then
    echo "Skipping Byzantine Tests because Docker Credentials were not present."
    BYZANTINE_IMAGE=""
else
    echo "$DOCKER_PASS" | docker login --username "$DOCKER_USERNAME" --password-stdin
    docker pull "${BYZANTINE_IMAGE}"
fi

E2E_TEST_COMMAND="${ROOT_DIRPATH}/scripts/build_and_run.sh"

# Docker only allows you to have spaces in the variable if you escape them or use a Docker env file
CUSTOM_ENV_VARS_JSON_ARG="CUSTOM_ENV_VARS_JSON={\"AVALANCHE_IMAGE\":\"${AVALANCHE_IMAGE}\",\"BYZANTINE_IMAGE\":\"${BYZANTINE_IMAGE}\"}"

return_code=0
if ! bash "${E2E_TEST_COMMAND}" all --env "${CUSTOM_ENV_VARS_JSON_ARG}" --env "PARALLELISM=${PARALLELISM}"; then
    echo "Avalanche E2E tests failed"
    return_code=1
else
    echo "Avalanche E2E tests succeeded"
    return_code=0
fi

exit "${return_code}"
