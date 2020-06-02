set -euo pipefail
SCRIPT_DIRPATH="$(cd "$(dirname "${0}")" && pwd)"
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"

DEFAULT_GECKO_IMAGE="kurtosistech/gecko:latest"
docker pull "${DEFAULT_GECKO_IMAGE}"

bash "${ROOT_DIRPATH}"/scripts/build_images.sh
#LATEST_INITIALIZER_TAG="kurtosistech/ava-e2e-tests_initializer:latest"
LATEST_CONTROLLER_TAG="kurtosistech/ava-e2e-tests_controller"

bash "${ROOT_DIRPATH}"/scripts/build.sh

("${ROOT_DIRPATH}"/build/ava-e2e-tests -gecko-image-name="${DEFAULT_GECKO_IMAGE}"\
 -test-controller-image-name="${LATEST_CONTROLLER_TAG}") &

#(docker run -v /var/run/docker.sock:/var/run/docker.sock \
#--env DEFAULT_GECKO_IMAGE="${DEFAULT_GECKO_IMAGE}" \
#--env TEST_CONTROLLER_IMAGE="${LATEST_CONTROLLER_TAG}" \
#"${LATEST_INITIALIZER_TAG}") &

kurtosis_pid="${!}"

# Sleep while Kurtosis spins up testnet and runs controller to execute tests.
sleep 90
docker image ls
docker ps -a
kill "${kurtosis_pid}"

ACTUAL_EXIT_STATUS="$(docker ps -a --latest --filter ancestor="${LATEST_CONTROLLER_TAG}" --format="{{.Status}}")"
EXPECTED_EXIT_STATUS="Exited \(0\).*"

echo "Exit status: ${ACTUAL_EXIT_STATUS}"

# Clear containers.
echo "Clearing kurtosis testnet containers."
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${DEFAULT_GECKO_IMAGE}" --format="{{.ID}}")) >/dev/null

if [[ ${ACTUAL_EXIT_STATUS} =~ ${EXPECTED_EXIT_STATUS} ]]
then
  echo "Kurtosis test succeeded."
  exit 0
else
  echo "Kurtosis test failed."
  exit 1
fi
