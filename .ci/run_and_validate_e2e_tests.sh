set -euo pipefail
SCRIPT_DIRPATH="$(cd "$(dirname "${0}")" && pwd)"
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"
PARALLELISM=4

BRANCH="${TRAVIS_BRANCH}"

DOCKER_REPO="avaplatform"
TESTING_REPO="$DOCKER_REPO/avalanche-testing"
AVALANCHE_REPO="$DOCKER_REPO/avalanchego"
DEFAULT_AVALANCHE_IMAGE="$DOCKER_REPO/avalanchego:dev"

function docker_tag_exists() {
    TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${DOCKER_USERNAME}'", "password": "'${DOCKER_PASS}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
    curl --silent -f --head -lL https://hub.docker.com/v2/repositories/$1/tags/$2/ > /dev/null
}

if docker_tag_exists $AVALANCHE_REPO $BRANCH; then
    echo "$AVALANCHE_REPO $BRANCH exists; using this image" 
    AVALANCHE_IMAGE="$AVALANCHE_REPO:$BRANCH"
else
    echo "$AVALANCHE_REPO $BRANCH does NOT exist; using the default image" 
    AVALANCHE_IMAGE=$DEFAULT_AVALANCHE_IMAGE
fi

echo "Using $AVALANCHE_IMAGE for CI"

# Use stable version of avalanche-byzantine based on everest for CI
BYZANTINE_IMAGE="$DOCKER_REPO/avalanche-byzantine:v0.1.4-rc.1"

# Kurtosis will try to pull Docker images, but as of 2020-08-09 it doesn't currently support pulling from Docker repos that require authentication
# so we have to do the pull here
docker pull "${AVALANCHE_IMAGE}"

# If Docker Credentials are not available skip the Byzantine Tests
if [[ -z ${DOCKER_USERNAME} ]]; then
    echo "Skipping Byzantine Tests because Docker Credentials were not present."
    BYZANTINE_IMAGE=""
else
    echo "$DOCKER_PASS" | docker login --username "$DOCKER_USERNAME" --password-stdin
    docker pull "${BYZANTINE_IMAGE}"
fi

echo "Build the image"
#build the image
AVALANCHE_TESTING_IMAGE=$TESTING_REPO
docker build -t $AVALANCHE_TESTING_IMAGE:$BRANCH . -f ${ROOT_DIRPATH}/testsuite/Dockerfile
#docker tag "$AVALANCHE_TESTING_IMAGE" "$BRANCH"
echo "$DOCKER_PASS" | docker login --username "$DOCKER_USERNAME" --password-stdin

# following should push all tags
echo "pushing image"
docker push $AVALANCHE_TESTING_IMAGE

echo "Starting build_and_run.sh"
E2E_TEST_COMMAND="AVALANCHE_IMAGE=$AVALANCHE_IMAGE ${ROOT_DIRPATH}/scripts/build_and_run.sh"

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
