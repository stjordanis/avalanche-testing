# Docker base image allows us to run docker-in-docker
FROM docker:stable

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    # Defaults for images point to DockerHub endpoints.
    # These should be set in docker container run args
    DEFAULT_GECKO_IMAGE="avaplatform/gecko" \
    TEST_CONTROLLER_IMAGE="kurtosistech/ava-e2e-tests:test-controller_latest"


# Move to working directory /build
WORKDIR /build

# Install Golang
RUN apk update && \
        apk upgrade && \
        apk add go

# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o test-initializer initializer/main.go

# Command to run when starting the container
CMD ./test-initializer --gecko-image-name=${DEFAULT_GECKO_IMAGE} --test-controller-image-name=${TEST_CONTROLLER_IMAGE}
