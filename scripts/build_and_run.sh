set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"

# ====================== CONSTANTS =======================================================
SUITE_IMAGE="avaplatform/avalanche-testing"
AVALANCHE_IMAGE="avaplatform/avalanchego:latest"
KURTOSIS_CORE_CHANNEL="master"
INITIALIZER_IMAGE="kurtosistech/kurtosis-core_initializer:${KURTOSIS_CORE_CHANNEL}"
API_IMAGE="kurtosistech/kurtosis-core_api:${KURTOSIS_CORE_CHANNEL}"

# ====================== ARG PARSING =======================================================
show_help() {
    echo "${0}:"
    echo "  -h      Displays this message"
    echo "  -b      Executes only the build step, skipping the run step"
    echo "  -r      Executes only the run step, skipping the build step"
    echo "  -d      Extra args to pass to 'docker run' (e.g. '--env MYVAR=somevalue')"
}

do_build=true
do_run=true
extra_docker_args=""
while getopts "brd:" opt; do
    case "${opt}" in
        h)
            show_help
            exit 0
            ;;
        b)
            do_run=false
            ;;
        r)
            do_build=false
            ;;
        d)
            extra_docker_args="${OPTARG}"
            ;;
    esac
done

# ====================== MAIN LOGIC =======================================================
git_branch="$(git rev-parse --abbrev-ref HEAD)"
docker_tag="$(echo "${git_branch}" | sed 's,[/:],_,g')"

root_dirpath="$(dirname "${script_dirpath}")"
if "${do_build}"; then
    echo "Running unit tests..."
    if ! go test "${root_dirpath}/..."; then
        echo "Tests failed!"
        exit 1
    else
        echo "Tests succeeded"
    fi

    echo "Building Avalanche testing suite image..."
    docker build -t "${SUITE_IMAGE}:${docker_tag}" -f "${root_dirpath}/testsuite/Dockerfile" "${root_dirpath}"
fi

if "${do_run}"; then
    suite_execution_volume="avalanche-test-suite_${docker_tag}_$(date +%s)"
    docker volume create "${suite_execution_volume}"

    custom_env_vars_json_flag="CUSTOM_ENV_VARS_JSON={\"AVALANCHE_IMAGE\":\"${AVALANCHE_IMAGE}\",\"BYZANTINE_IMAGE\":\"\"}"
    echo "${custom_env_vars_json_flag}"
    docker run \
        --mount "type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock" \
        --mount "type=volume,source=${suite_execution_volume},target=/suite-execution" \
        --env "${custom_env_vars_json_flag}" \
        --env "TEST_SUITE_IMAGE=${SUITE_IMAGE}:${docker_tag}" \
        --env "SUITE_EXECUTION_VOLUME=${suite_execution_volume}" \
        --env "KURTOSIS_API_IMAGE=${API_IMAGE}" \
        ${extra_docker_args} \
        "${INITIALIZER_IMAGE}"
fi
