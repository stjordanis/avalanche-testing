# TBD
* Use Kurtosis version that allows the user to configure network width
* Use Kurtosis version where `TestSuiteRunner.RunTests` takes in a set of tests, rather than a list
* Parameterize the `GeckoService` struct with ports so that it's not implicitly relying on constants from `GeckoServiceInitializerCore`

# 0.5.0
* Make the fully-connected-node test actually test staker registration by ensuring a second node sees the newly-registered-as-staker node
* Upgrade to Kurtosis version with hard test timeouts, to prevent infinite hangs
* Added an unrequested chit spammer Byzantine test
* Increasing timeouts in startup and in duplicate ID test to work in gecko CI
* Add logic to wait for addition of validator in default subnet list

# 0.4.0
* Drop default loglevel for initializer & controller down to DEBUG
* Upgrade controller Docker image to allow for a Docker network per test
* Run tests in parallel!
* Fix bug in RPC workflow test where delegator ID was actually staker ID
* Removed references to `fiveNode` in testsuite (because these networks are no longer five-node, and the network size is less important than whether it's staking or not)
* Removed `fiveStakingNodeGetValidatorsTest`, which has been superseded by the RPC workflow test
* Provided `--list` flag to the initializer CLI to simply list the tests registered in the suite
* Switched `int` ports to `nat.Port` to allow for specifying non-TCP ports

# 0.3.1
* Specify --http-host CLI flag in GetStartCommand to have RPC calls bind to publicIP
* Migrate AdminAPI endpoints to the new InfoAPI
* Parse error message on XChain Imports in order to wait for PChain transaction to be accepted
* Added tests for XChain endpoints
* Remove `FiveNodeStakingNetworkBasicTest` (wasn't being used)
* Test if the network functions as expected when nodes with duplicate node IDs occur

# 0.3.0
* Removed startPortRange and endPortRange CLI args
* Declare timeouts for tests
* After reviewing possibility of centralizing deserialization of JSON RPC responses, determined it's not doable. Removing TODOs related to this.
* Make initializer & controller log levels independently configurable
* Adding test for full connectivity based on Peers API.
* Split `admin` and `health` endpoint types into their own files, for readability
* Provide functionality for tests to add & stop nodes in the network dynamically
* Split Gecko networks into MutableGeckoNetwork and FixedGeckoNetwork
* Fix bug with tests passing when they shouldn't
* Catch RPC-level errors in MakeRPCRequest
* Add all five default stakers to staking network bootstrapping
* Implement test for transferring assets between XChain accounts
* Implement test for transferring assets from XChain to PChain
* Created `GeckoCertProvider` interface that's fed into the `GeckoServiceInitializerCore`, allowing for test writers to customize the certs that the certs get
* Created two implementations of `GeckoCertProvider`:
    * `StaticGeckoCertProvider`, which provides the exact same predefined cert repeatedly
    * `RandomGeckoCertProvider`, which provides generated certs (optionally, the same random-generated cert each time)
* Removed `FixedGeckoNetwork` in favor of `TestGeckoNetwork`, which allows for more control over the testnet that gets created
* Removed the single-node and ten-node Gecko tests; they don't actually test anything useful when compared to the staking network tests
* Expanded RpcWorkflow test to add a staker
* Expanded RpcWorkflow test to add a delegator and transfer funds back to XChain
* Created high level function to both fund and add a staker to default subnet
* Fixed fully connected test and added nonbootstrap node as staker

# 0.2.1
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
* Removed unnecessary docker pull command in ci script.

# 0.2.0
* Updated code to use latest Kurtosis version that stops containers
* Added convenience scripts for rebuilding and running to the `scripts` directory
* Updated CI check & documentation to reflect the use of the convenience scripts
* Switched ServiceAvailabilityCheckerCore to use the `health` API's getLiveness call
* Added `GeckoClient` with some basic endpoint implementations
* Enabled staking
* Added grabbing logs from the controller
