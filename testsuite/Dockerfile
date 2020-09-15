FROM golang:1.13-alpine
WORKDIR /build
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o avalanche-test-suite testsuite/main.go

# TODO Get rid of tee/LOG_FILEPATH in favor of using a Docker logging driver in the initializer
CMD set -euo pipefail && ./avalanche-test-suite \
    --metadata-filepath=${METADATA_FILEPATH} \
    --test=${TEST} \
    --log-level=${LOG_LEVEL} \
    --services-relative-dirpath=${SERVICES_RELATIVE_DIRPATH} \
    --avalanche-go-image=${AVALANCHE_IMAGE} \
    --byzantine-go-image=${BYZANTINE_IMAGE} \
    --kurtosis-api-ip=${KURTOSIS_API_IP} 2>&1 | tee ${LOG_FILEPATH}
