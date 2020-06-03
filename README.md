# Ava End-to-End Tests
This repo contains end-to-end tests for the Ava network using [the Kurtosis testing framework](https://github.com/kurtosis-tech/kurtosis)

## Requirements
* Golang version 1.13x.x
* A Docker engine running in your environment

## Local Install
1. Clone this repository
1. Run `scripts/build_controller_image.sh`
1. Run `scripts/build.sh`
1. Run `build/ava-e2e-tests --help` to see the available flags for running the CLI
1. Run the binary with the desired flags, noting that:
    * The Gecko image argument must be a Docker image built for Gecko
    * The test controller image argument will likely be `kurtosistech/ava-e2e-tests_controller:latest` (which was created when you built the image above)

**A helpful tip:** as of 2020-06-02, the Kurtosis library doesn't yet stop the Docker containers that got started. You can use the following alias in your `.bashrc` to clear the running containers:

```
# alias for clearing kurtosis containers 
clear_containers() {  docker rm $(docker stop $(docker ps -a -q --filter ancestor="$1" --format="{{.ID}}")); } 
alias cclear=clear_containers
```

Usage:
```
export GECKO_IMAGE=gecko-684ca4e
export CONTROLLER_IMAGE=kurtosistech/ava-e2e-tests_controller
# run the tests
./build/ava-e2e-tests -gecko-image-name="${GECKO_IMAGE}" -test-controller-image-name="${CONTROLLER_IMAGE}"
# ...Ctrl-C to kill the test CLI...
# clear the docker containers initialized by the tests
cclear ${GECKO_IMAGE} 
cclear ${CONTROLLER_IMAGE} 
```
