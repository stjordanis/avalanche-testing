# Ava End-to-End Tests
This repo contains end-to-end tests for the Ava network using [the Kurtosis testing framework](https://github.com/kurtosis-tech/kurtosis)

## Requirements
* Golang version 1.13x.x
* A Docker engine running in your environment

## Local Install
1. Clone this repository
1. Run `scripts/build_images.sh`
1. Run 
2. Run `./scripts/build.sh`. This will build the main binary and put it the `build/` directory of this repository.  

Clone [the test controller](https://github.com/kurtosis-tech/ava-test-controller) and run `docker build .` inside the directory.


## TODO
This repo is designed to run inside a Docker container started by the [Kurtosis](https://github.com/kurtosis-tech/kurtosis) testing tool.

## Helpful Tip

Create an alias in your shell .rc file to stop and clear all Docker containers created by Kurtosis in one line.  
Run this every time after you kill kurtosis, because the containers will hang around.  
One way to do this is as follows:

```
# alias for clearing kurtosis containers 
kurtosisclearall() {  docker rm $(docker stop $(docker ps -a -q --filter ancestor="$1" --format="{{.ID}}")) } 
alias kclear=kurtosisclearall
```

Usage:
```
export GECKO_IMAGE=gecko-684ca4e
# run kurtosis
./build/kurtosis -gecko-image-name="${GECKO_IMAGE}"
# ...kill kurtosis manually...
# clear the docker containers initialized by kurtosis
kclear ${GECKO_IMAGE} 
```


# Kurtosis in Docker

Kurtosis in Docker is based on the "Docker in Docker" docker image.
It connects to your host docker engine, rather than deploying its own docker engine.
This means Kurtosis will be running in a container in the same docker environment as the testnet and test controller containers.

### Running Kurtosis in Docker

In the root directory of this repository, run 
`./scripts/build_image.sh` to build the Kurtosis docker image. It will create an image with tag kurtosis-<COMMIT_HASH>.

To run Kurtosis in Docker, be sure to bind the docker socket of the container with the host docker socket, so they use the same docker engine.
Also, specify the Gecko image and the Test Controller image at container runtime.

Example command:

`docker run -ti -v /var/run/docker.sock:/var/run/docker.sock --env DEFAULT_GECKO_IMAGE="gecko-60668c3" --env TEST_CONTROLLER_IMAGE=ava-controller:latest kurtosis-5006918`
