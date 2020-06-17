set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"

GECKO_IMAGE_DEFAULT="kurtosistech/gecko"
CONTROLLER_IMAGE="kurtosistech/ava-e2e-tests_controller:latest"
root_dirpath="$(dirname "${script_dirpath}")"

"${root_dirpath}/build/ava-e2e-tests" "--gecko-image-name=${GECKO_IMAGE_DEFAULT}" "--test-controller-image-name=${CONTROLLER_IMAGE}" ${*:-}
