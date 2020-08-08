set -euo pipefail
SCRIPT_DIRPATH="$(cd "$(dirname "${0}")" && pwd)"
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"

# login to AWS for byzantine images
aws ecr get-login-password --region "${AWS_DEFAULT_REGION}" | docker login --username AWS --password-stdin 964377072876.dkr.ecr.us-east-1.amazonaws.com

DEFAULT_CONTROLLER_TAG="kurtosistech/ava-e2e-tests_controller"   # TODO This is hardcoded in the full_rebuild_and_run.sh script - this should be parameterized!!!!
DEFAULT_GECKO_IMAGE="964377072876.dkr.ecr.us-east-1.amazonaws.com/gecko:latest"
BYZANTINE_IMAGE="964377072876.dkr.ecr.us-east-1.amazonaws.com/gecko-byzantine:latest"

docker pull "${BYZANTINE_IMAGE}"

E2E_TEST_COMMAND="${ROOT_DIRPATH}/scripts/full_rebuild_and_run.sh"
BYZANTINE_IMAGE_ARG="--byzantine-image-name=${BYZANTINE_IMAGE}"
GECKO_IMAGE_ARG="--gecko-image-name=${DEFAULT_GECKO_IMAGE}"
return_code=0
if ! bash "${E2E_TEST_COMMAND}" "${BYZANTINE_IMAGE_ARG}" "${DEFAULT_GECKO_IMAGE}"; then
    echo "Ava E2E tests failed"
    return_code=1
else
    echo "Ava E2E tests succeeded"
    return_code=0
fi

# Clear containers.
echo "Clearing Ava Docker containers..."
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${DEFAULT_GECKO_IMAGE}" --format="{{.ID}}")) >/dev/null
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${BYZANTINE_IMAGE}" --format="{{.ID}}")) >/dev/null
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${DEFAULT_CONTROLLER_TAG}" --format="{{.ID}}")) >/dev/null
echo "Ava Docker containers cleared successfully"

exit "${return_code}"
