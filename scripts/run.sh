set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"

# This Docker image defines the gecko image used for developing e2e tests.
GECKO_IMAGE_DEFAULT="avaplatform/gecko"
CONTROLLER_IMAGE="avaplatform/avalanche-e2e-tests_controller:latest"
root_dirpath="$(dirname "${script_dirpath}")"

"${root_dirpath}/build/avalanche-e2e-tests" "--gecko-image-name=${GECKO_IMAGE_DEFAULT}" "--test-controller-image-name=${CONTROLLER_IMAGE}" ${*:-}
