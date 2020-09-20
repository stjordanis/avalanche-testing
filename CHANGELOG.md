# TBD
* Upgraded to Kurtosis 1.0
* Centralized run scripts into `build_and_run.sh`

# 0.9.0
* Update to v0.7.0 of avalanchego and avalanche-byzantine
* Rename delegator/staker functions
* Add conflicting transactions vertex test
* Switch to using AvaLabs Docker registry for avalanchego and avalanche-byzantine images
* Add docs to every public function and struct
* Update avalanche client to use the same structs used by avalanchego API services
* Fix bug in RPCWorkFlowTest where multiple clients shared the genesis key leading to undefined behavior
* Refactor code organization and comments
* Point CI to latest everest images for both avalanchego and avalanche-byzantine
* Add bombard test to bombard the X chain with transactions and then add two nodes to bootstrap the new data
* Update get current and pending validators calls in platform API
* Rename from gecko -> avalanchego

# 0.8.0
* Switch configuration IDs to strings instead of ints
* Bump kurtosis version to get cleanup on ctrl-c
* Pull in Kurtosis version that will print test outputs they finish, rather than waiting for all tests to finish
* Point CI to stable Docker image based off of Denali for both gecko and gecko-byzantine

# 0.7.0
* Split `staking_network_tests` into separate files per test
* Upgrade to Kurtosis version with simplified service config definition
* Implement network consensus timeouts in high level gecko client as fractions of total test timeout
* Add CI checks to make sure changelog is updated
* Upgrade to Kurtosis version using custom structs for service/config IDs (rather than ints)
* Significantly up test execution timeouts
* Make controller Docker image `tee` to the logfile, rather than redirecting all output
* Implement a 10-second HTTP request timeout in the Gecko client
* Upgraded to Kurtosis that allows the setup buffer to be configured at a per-test level, and gave a generous setup buffer for as long as the Gecko availability checker core has a `time.Sleep` in it
* Rename HighLevelGeckoClient to RPCWorkFlowRunner and move to ava_testsuite package
* Bugfix on Travis CI checks for CHANGELOG entries
* Add comments and documentation in RPC workflow and chit spammer tests
* Using network ID in controller image rather than network name
* Changing service IDs to strings rather than integers
* Update function interface for GetStartCommand according to net.IP change in kurtosis
* Added a script to clean Docker environment

# 0.6.0
* Use Kurtosis version that allows the user to configure network width
* Use Kurtosis version where `TestSuiteRunner.RunTests` takes in a set of tests, rather than a list
* Move `ServiceSocket`, which is Ava-specific, to this repo from Kurtosis
* Parameterize the `GeckoService` struct with ports so that it's not implicitly relying on constants from `GeckoServiceInitializerCore`
* Fix breaks caused by small Kurtosis cleanups
* Rework the README to do a better job explaining what this repo contains
* Increasing timeouts in startup and in duplicate ID test to work in gecko CI
* Add logic to wait for addition of validator in default subnet list

# 0.5.0
* Make the fully-connected-node test actually test staker registration by ensuring a second node sees the newly-registered-as-staker node
* Upgrade to Kurtosis version with hard test timeouts, to prevent infinite hangs
* Added an unrequested chit spammer Byzantine test

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
