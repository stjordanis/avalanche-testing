FROM golang:1.13-alpine

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["controller", "${TEST_NAME}", "${NETWORK_DATA_FILEPATH}"]
