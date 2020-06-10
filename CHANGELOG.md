# TBD
* Fix tee suppressing exit code of the Docker image
* Add parameters to `makeRpcRequest` on Gecko client
* Add `GetBlockchainStatus` endpoint on Gecko client
* Added a mock JSON RPC requester for testing Gecko client methods
* Used the mock requester to write tests for `GetCurrentValidators` and `GetBlockchainStatus`

# 0.2.0
* Updated code to use latest Kurtosis version that stops containers
* Added convenience scripts for rebuilding and running to the `scripts` directory
* Updated CI check & documentation to reflect the use of the convenience scripts
* Switched ServiceAvailabilityCheckerCore to use the `health` API's getLiveness call
* Added `GeckoClient` with some basic endpoint implementations
* Enabled staking
* Added grabbing logs from the controller
