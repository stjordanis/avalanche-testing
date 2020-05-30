FROM golang:1.13-alpine

WORKDIR /build
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o controller .

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist
ENV PATH="/dist:${PATH}"

# Copy binary from build to main folder
RUN cp /build/controller .

CMD ["controller"]

