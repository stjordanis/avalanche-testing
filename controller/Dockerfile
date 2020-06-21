ENV binary test-controller
ENV buildWorkdir /build

# ============= Build Stage ======================
FROM golang:1.13-alpine AS builder
WORKDIR ${buildWorkdir}
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .

# Build the application
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ${binary} controller/main.go

# ============= Execution Stage ================
FROM docker:stable AS execution
WORKDIR /run

# Copy the code into the container
COPY --from=builder ${buildWorkdir}/${binary} .

# Note that this CANNOT be an execution list else the variables won't be expanded
# See: https://stackoverflow.com/questions/40454470/how-can-i-use-a-variable-inside-a-dockerfile-cmd
CMD ./test-controller \
    --test=${TEST_NAME} \
    --test-image-name=${TEST_IMAGE_NAME} \
    --subnet-mask=${SUBNET_MASK} \
    --test-controller-ip=${TEST_CONTROLLER_IP} \
    --gateway-ip=${GATEWAY_IP} \
    --log-level=${LOG_LEVEL} &> ${LOG_FILEPATH}
