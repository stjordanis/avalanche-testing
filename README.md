Avalanche End-to-End Tests
====================
This repo contains end-to-end tests for the Avalanche network and Gecko client using [the Kurtosis testing framework](https://github.com/kurtosis-tech/kurtosis)

* [Requirements](#requirements)
* [Running Locally](#running-locally)
* [Developing Locally](#developing-locally)
    * [Architecture](#architecture)
    * [Adding A Test](#adding-a-test)
    * [Running Locally As A Developer](#running-locally-as-a-developer)
    * [Keeping Your Dev Environment Clean](#keeping-your-dev-environment-clean)

Requirements
------------
* Golang version 1.13x.x
* [A Docker engine running in your environment](https://docs.docker.com/engine/install/)

Running Locally
---------------
1. Clone this repository
1. Run `scripts/full_rebuild_and_run.sh` and wait for it to finish (which will take a while)

This command will spin up multiple Docker containers during its operation, and you can examine container logs via either your Docker engine dashboard GUI or the `docker container ls` and `docker container logs`.

NOTE: The Avalanche E2E test suite defaults to running 4 tests in parallel to speed up test suite execution time. If your machine has less cores, you should reduce this parallelism to _at maximum_ the number of cores on your machine, else the extra context-switching will slow down test execution and potentially cause spurious failures. To set the paralleism, pass the `--parallelism=N` argument to `full_rebuild_and_run.sh` (where "N" is the desired number of threads).

Once `full_rebuild_and_run.sh` has finished, you can now execute `scripts/run.sh` to re-run the testing suite without needing to rebuild. `run.sh` will accept arguments to modify test suite execution; to see the full list of supported arguments, pass in the `--help` flag.

Developing Locally
------------------
This repo uses the [Kurtosis architecture](https://github.com/kurtosis-tech/kurtosis), so you should first go through the tutorial there to familiarize yourself with the core Kurtosis concepts.

In this implementation of Kurtosis, we have:
* `AvalancheService` interface to represent the actions a test can take against a  generic service participating in the Avalanche network being tested
* `GeckoService` interface to represent the actions a test can take against a Gecko Avalanche client participating in the Avalanche network being tested
* `GeckoServiceInitializerCore` and `GeckoServiceAvailabilityChecker` for instantiating Gecko Ava clients in test networks
    * `GeckoCertProvider` to allow controlling the cert that a Gecko node starts with, to allow for writing duplicate-node-ID tests
* `TestGeckoNetwork` to encapsulate a test Avalanche network of Gecko nodes of arbitrary size
* Several tests
* `AvalancheTestSuite` to contain all the tests Kurtosis can run
* A `main.go` for running a controller Docker image under the `controller` package
* A `main.go` for running the Kurtosis initializer under the `initializer` package

Additionally, for ease of writing tests, this repo also contains a Go client for interacting with the JSON RPC API of a Gecko service (which should probably be moved to the Gecko repo).

### Adding A Test
1. Create a new file in `commons/testsuite` for your test
1. Create a struct that implements the `testsuite.Test` interface from Kurtosis
1. Fill in the interface's functions
1. Register the test in `AvalancheTestSuite`'s `GetTests` method

### Running Your Code
The `scripts/full_rebuild_and_run.sh` will rebuild and rerun both the initializer and controller Docker image; rerun this every time that you make a change. Arguments passed to this script will get passed to the initializer binary CLI as-is.

### Keeping Your Dev Environment Clean
Kurtosis intentionally doesn't delete containers and volumes, which means your local Docker environment will accumulate images, containers, and volumes; you can use [the script here](./scripts/clean_docker_environment.sh) to clean old containers and images. For further information, read [the Notes section of the Kurtosis README](https://github.com/kurtosis-tech/kurtosis/tree/develop#notes) for more details on how to keep your local environment clean while you develop.
