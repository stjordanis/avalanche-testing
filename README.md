# Ava Test Controller

This repo is designed to run inside a Docker container started by the [Kurtosis](https://github.com/kurtosis-tech/kurtosis) testing tool.

Kurtosis will provide two environment variables, `TEST_NAME` and `NETWORK_DATA_FILEPATH` which will contain, respectively, 1) the name of the test to run and 
2) the filepath to a JSON-serialized representation of the Docker network that the test should connect to.
