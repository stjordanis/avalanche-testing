package connected

import (
	"time"

	avalancheNetwork "github.com/ava-labs/avalanche-e2e-tests/avalanche/networks"
	avalancheService "github.com/ava-labs/avalanche-e2e-tests/avalanche/services"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/avalanche-e2e-tests/testsuite/helpers"
	"github.com/ava-labs/avalanche-e2e-tests/testsuite/verifier"
	"github.com/ava-labs/gecko/api"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
)

const (
	stakerUsername = "staker"
	stakerPassword = "test34test!23"
	seedAmount     = uint64(50000000000000)
	stakeAmount    = uint64(30000000000000)

	normalNodeConfigID networks.ConfigurationID = "normal-config"

	networkAcceptanceTimeoutRatio                    = 0.3
	nonBootValidatorServiceID     networks.ServiceID = "validator-service"
	nonBootNonValidatorServiceID  networks.ServiceID = "non-validator-service"
)

// StakingNetworkFullyConnectedTest adds nodes to the network and verifies that the network stays fully connected
type StakingNetworkFullyConnectedTest struct {
	ImageName string
	Verifier  verifier.NetworkStateVerifier
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkFullyConnectedTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(avalancheNetwork.TestGeckoNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))

	stakerIDs := castedNetwork.GetAllBootServiceIDs()
	allServiceIDs := make(map[networks.ServiceID]bool)
	for stakerID := range stakerIDs {
		allServiceIDs[stakerID] = true
	}
	// Add our custom nodes
	allServiceIDs[nonBootValidatorServiceID] = true
	allServiceIDs[nonBootNonValidatorServiceID] = true

	allNodeIDs, allGeckoClients := getNodeIDsAndClients(context, castedNetwork, allServiceIDs)
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIDs, stakerIDs, allNodeIDs, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}

	nonBootValidatorClient := allGeckoClients[nonBootValidatorServiceID]
	highLevelExtraStakerClient := helpers.NewRPCWorkFlowRunner(
		nonBootValidatorClient,
		api.UserPass{Username: stakerUsername, Password: stakerPassword},
		networkAcceptanceTimeout)
	if _, err := highLevelExtraStakerClient.ImportGenesisFundsAndStartValidating(seedAmount, stakeAmount); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add extra staker."))
	}

	// Give time for the new validator to propagate via gossip
	time.Sleep(70 * time.Second)

	stakerIDs[nonBootValidatorServiceID] = true

	/*
		After gossip, we expect the peers list to look like:
		1) No node has itself in its peers list
		2) The validators will have ALL other nodes in the network (propagated via gossip)
		3) The non-validators will have all the validators in the network (propagated via gossip)
	*/
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIDs, stakerIDs, allNodeIDs, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying that the network is fully connected after gossip"))
	}
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkFullyConnectedTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	serviceConfigs := map[networks.ConfigurationID]avalancheNetwork.TestGeckoNetworkServiceConfig{
		normalNodeConfigID: *avalancheNetwork.NewTestGeckoNetworkServiceConfig(
			true,
			avalancheService.DEBUG,
			test.ImageName,
			2,
			2,
			make(map[string]string),
		),
	}
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{
		nonBootValidatorServiceID:    normalNodeConfigID,
		nonBootNonValidatorServiceID: normalNodeConfigID,
	}
	return avalancheNetwork.NewTestGeckoNetworkLoader(
		true,
		test.ImageName,
		avalancheService.DEBUG,
		2,
		2,
		0,
		serviceConfigs,
		desiredServices,
	)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test StakingNetworkFullyConnectedTest) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test StakingNetworkFullyConnectedTest) GetSetupBuffer() time.Duration {
	// TODO drop this when the availabilityChecker doesn't have a sleep (because we spin up a bunch of nodes before running the test)
	return 6 * time.Minute
}

// ================ Helper functions =========================
/*
This helper function will grab node IDs and Gecko clients
*/
func getNodeIDsAndClients(
	testContext testsuite.TestContext,
	network avalancheNetwork.TestGeckoNetwork,
	allServiceIDs map[networks.ServiceID]bool,
) (allNodeIDs map[networks.ServiceID]string, allGeckoClients map[networks.ServiceID]*apis.Client) {
	allGeckoClients = make(map[networks.ServiceID]*apis.Client)
	allNodeIDs = make(map[networks.ServiceID]string)
	for serviceID := range allServiceIDs {
		client, err := network.GetAvalancheClient(serviceID)
		if err != nil {
			testContext.Fatal(stacktrace.Propagate(err, "An error occurred getting the Gecko client for service with ID %v", serviceID))
		}
		allGeckoClients[serviceID] = client
		nodeID, err := client.InfoAPI().GetNodeID()

		if err != nil {
			testContext.Fatal(stacktrace.Propagate(err, "An error occurred getting the Gecko node ID for service with ID %v", serviceID))
		}
		allNodeIDs[serviceID] = nodeID
	}
	return
}
