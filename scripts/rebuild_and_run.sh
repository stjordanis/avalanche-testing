set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"

bash "${script_dirpath}/build_controller_image.sh"
bash "${script_dirpath}/build.sh"

# Allow user to override default Gecko image name if desired
if [ "${#}" -gt 0 ]; then
    bash "${script_dirpath}/run.sh" "${1}"
else
    bash "${script_dirpath}/run.sh"
fi
