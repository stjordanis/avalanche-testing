set -euo pipefail
SCRIPT_DIRPATH="$(cd "$(dirname "${0}")" && pwd)"
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"

AWS_ACCESS_KEY_ID="${BYZ_REG_AWS_ID}"
AWS_SECRET_ACCESS_KEY="${BYZ_REG_AWS_KEY}"
AWS_DEFAULT_REGION="${BYZ_REG_AWS_REGION}"
# login to AWS for byzantine images
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 964377072876.dkr.ecr.us-east-1.amazonaws.com

DEFAULT_CONTROLLER_TAG="kurtosistech/ava-e2e-tests_controller"
DEFAULT_GECKO_IMAGE="kurtosistech/gecko:latest"

return_code=0
if ! bash "${ROOT_DIRPATH}/scripts/full_rebuild_and_run.sh"; then
    echo "Ava E2E tests failed"
    return_code=1
else
    echo "Ava E2E tests succeeded"
    return_code=0
fi

# Clear containers.
echo "Clearing Ava Docker containers..."
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${DEFAULT_GECKO_IMAGE}" --format="{{.ID}}")) >/dev/null
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${DEFAULT_CONTROLLER_TAG}" --format="{{.ID}}")) >/dev/null
echo "Ava Docker containers cleared successfully"

exit "${return_code}"
