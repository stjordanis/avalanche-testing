set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"

# This Docker image defines the avalanche image used for developing e2e tests.
AVALANCHE_IMAGE_DEFAULT="avaplatform/avalanchego:latest"
CONTROLLER_IMAGE="avaplatform/avalanche-testing_controller:latest"
root_dirpath="$(dirname "${script_dirpath}")"

"${root_dirpath}/build/avalanche-testing" "--avalanche-image-name=${AVALANCHE_IMAGE_DEFAULT}" "--test-controller-image-name=${CONTROLLER_IMAGE}" ${*:-}
