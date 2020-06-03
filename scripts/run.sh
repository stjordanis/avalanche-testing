set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"

GECKO_IMAGE_DEFAULT="kurtosistech/gecko"
CONTROLLER_IMAGE="kurtosistech/ava-e2e-tests_controller:latest"
root_dirpath="$(dirname "${script_dirpath}")"

# Allow user to override default Gecko image name if desired
gecko_image="${GECKO_IMAGE_DEFAULT}"
if [ "${#}" -gt 0 ]; then
    gecko_image="${1}"
fi

"${root_dirpath}/build/ava-e2e-tests" "--gecko-image-name=${gecko_image}" "--test-controller-image-name=${CONTROLLER_IMAGE}"
