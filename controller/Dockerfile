FROM golang:1.13-alpine
FROM docker:stable

WORKDIR /build
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o test-controller controller/main.go

# Note that this CANNOT be an execution list else the variables won't be expanded
# See: https://stackoverflow.com/questions/40454470/how-can-i-use-a-variable-inside-a-dockerfile-cmd
CMD ./test-controller \
    --test=${TEST_NAME} \
    --test-image-name=${TEST_IMAGE_NAME} \
    --subnet-mask=${SUBNET_MASK} \
    --test-controller-ip=${TEST_CONTROLLER_IP} \
    --gateway-ip=${GATEWAY_IP} \
    --log-level=${LOG_LEVEL} &> ${LOG_FILEPATH}
