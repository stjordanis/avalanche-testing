# TBD
* Fix tee suppressing exit code of the Docker image
* Add parameters to `makeRpcRequest` on Gecko client
* Added a mock JSON RPC requester for testing Gecko client methods
* Add method calls on the Gecko client and tests for the following PChain endpoints:
    * `createBlockhain`
    * `getBlockchainStatus`
    * `createAccount`
    * `importKey`
    * `exportKey`
    * `getAccount`
    * `listAccounts`
    * `createSubnet`
    * `platform.getSubnets`
    * `platform.validatedBy`
    * `platform.validates`
    * `platform.getBlockchains`
    * `platform.exportAVA`
    * `platform.importAVA`
    * `platform.sign`
    * `platform.issueTx`
    * `getPendingValidators`
    * `sampleValidators`
    * `addDefaultSubnetValidator`
    * `addNonDefaultSubnetValidator`
    * `addNonDefaultSubnetDelegator`
* Added tests for the following non-PChain endpoints:
    * `admin.getNodeID`
    * `admin.peers`
    * `health.getLiveness`

# 0.2.0
* Updated code to use latest Kurtosis version that stops containers
* Added convenience scripts for rebuilding and running to the `scripts` directory
* Updated CI check & documentation to reflect the use of the convenience scripts
* Switched ServiceAvailabilityCheckerCore to use the `health` API's getLiveness call
* Added `GeckoClient` with some basic endpoint implementations
* Enabled staking
* Added grabbing logs from the controller
