set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"

GECKO_IMAGE_DEFAULT="kurtosistech/gecko"
CONTROLLER_IMAGE="kurtosistech/ava-e2e-tests_controller:latest"
root_dirpath="$(dirname "${script_dirpath}")"

test_names="${1:-}"

"${root_dirpath}/build/ava-e2e-tests" "--gecko-image-name=${GECKO_IMAGE_DEFAULT}" "--test-controller-image-name=${CONTROLLER_IMAGE}" --test-names="${test_names}"
