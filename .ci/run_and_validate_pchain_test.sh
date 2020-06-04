set -euo pipefail
SCRIPT_DIRPATH="$(cd "$(dirname "${0}")" && pwd)"
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"

DEFAULT_CONTROLLER_TAG="kurtosistech/ava-e2e-tests_controller"

# TODO do we even need to pull here? Kurtosis should pull this image automatically
DEFAULT_GECKO_IMAGE="kurtosistech/gecko:latest"
docker pull "${DEFAULT_GECKO_IMAGE}"

bash "${ROOT_DIRPATH}/scripts/full_rebuild_and_run.sh" "tenNodeGetValidatorsTest"
ACTUAL_EXIT_STATUS="$(docker ps -a --latest --filter ancestor="${DEFAULT_CONTROLLER_TAG}" --format="{{.Status}}")"
EXPECTED_EXIT_STATUS="Exited \(0\).*"

echo "Exit status: ${ACTUAL_EXIT_STATUS}"

# Clear containers.
echo "Clearing Ava Docker containers..."
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${DEFAULT_GECKO_IMAGE}" --format="{{.ID}}")) >/dev/null
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${DEFAULT_CONTROLLER_TAG}" --format="{{.ID}}")) >/dev/null
echo "Ava Docker containers cleared successfully"

if [[ ${ACTUAL_EXIT_STATUS} =~ ${EXPECTED_EXIT_STATUS} ]]
then
  echo "Ava E2E tests succeeded."
  exit 0
else
  echo "Ava E2E tests failed."
  exit 1
fi
