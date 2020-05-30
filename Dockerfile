FROM golang:1.13-alpine

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

# Note that this CANNOT be an execution list else the variables won't be expanded
# See: https://stackoverflow.com/questions/40454470/how-can-i-use-a-variable-inside-a-dockerfile-cmd
CMD controller --test=$TEST_NAME --network-info-filepath=$NETWORK_DATA_FILEPATH
