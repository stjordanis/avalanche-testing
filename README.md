# Ava End-to-End Tests
This repo contains end-to-end tests for the Ava network using [the Kurtosis testing framework](https://github.com/kurtosis-tech/kurtosis)

## Requirements
* Golang version 1.13x.x
* [A Docker engine running in your environment](https://docs.docker.com/engine/install/)

## Running Locally
1. Clone this repository
1. Run `scripts/build_controller_image.sh`
1. Run `scripts/build.sh`
1. Run `build/ava-e2e-tests --help` to see the available flags for running the CLI
1. Run the binary with the desired flags, noting that:
    * The Gecko image argument must be a Docker image built for Gecko
    * The test controller image argument will likely be `kurtosistech/ava-e2e-tests_controller:latest` (which was created when you built the image above)
1. When you see output like `INFO[0006] Waiting for containerId xxxxxx` then you can Ctl-C to kill the CLI (auto-ending coming this week)

In your Docker engine you'll now see:
* A set of running Docker containers representing the nodes of the test networks (which aren't yet cleaned up - also coming this week)
* A set of stopped Docker containers for the test controllers, one for each test

To view the results of your tests, open the logs of the stopped test controller containers (this will also be improved this week)

### Helpful Tip
You can use the following alias to clear running & stopped Docker containers of a certain type:

```
# alias for clearing kurtosis containers 
clear_containers() {  docker rm $(docker stop $(docker ps -a -q --filter ancestor="$1" --format="{{.ID}}")); } 
alias cclear=clear_containers
```

Example for clearing the test nodes and test controllers:
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

## Writing Tests
This repo uses the [Kurtosis architecture](https://github.com/kurtosis-tech/kurtosis), so you'll want to be familiar with the concepts there. In this implementation:
* The `AvaTestSuite` struct defines the tests that will be run
* The `AvaTestSuite` struct gets registered with the initializer and the controller in `initializer/main.go` and `controller/main.go` respectively
* The tests, the networks the tests will run against, and the services the networks are composed of live in the `commons` package

### Adding A Test
1. Create a new file in `commons/ava_testsuite` for your test
1. Create a struct that implements the `testsuite.Test` interface from Kurtosis
1. Fill in:
    1. The function defining which network the test will use
    1. The test logic
1. Add the test to the `AvaTestSuite`'s `GetTests` method

### Adding A Network
1. Create a new file in `commons/ava_networks` for your network
1. Create a struct representing the network and the calls a test could make against the network (e.g. `GetNodeX(i int)`)
1. Create a struct implementing `TestNetworkLoader` with the methods that will:
    1. For the initializer, configure your network using `GetNetworkConfig`
    1. For the controller, create your struct from node IP information
1. Configure tests (or write new ones) to use your network's loader
