package connected

import (
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	avalancheNetwork "github.com/ava-labs/avalanche-testing/avalanche/networks"
	avalancheService "github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/testsuite/helpers"
	"github.com/ava-labs/avalanche-testing/testsuite/verifier"
	"github.com/ava-labs/avalanchego/api"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
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
	ImageName           string
	FullyConnectedDelay time.Duration
	Verifier            verifier.NetworkStateVerifier
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkFullyConnectedTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(avalancheNetwork.TestAvalancheNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))

	stakerIDs := castedNetwork.GetAllBootServiceIDs()
	allServiceIDs := make(map[networks.ServiceID]bool)
	for stakerID := range stakerIDs {
		allServiceIDs[stakerID] = true
	}
	// Add our custom nodes
	allServiceIDs[nonBootValidatorServiceID] = true
	allServiceIDs[nonBootNonValidatorServiceID] = true

	allNodeIDs, allAvalancheClients := getNodeIDsAndClients(context, castedNetwork, allServiceIDs)
	logrus.Infof("Verifying that the network is fully connected...")
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIDs, stakerIDs, allNodeIDs, allAvalancheClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}
	logrus.Infof("Network is fully connected.")

	logrus.Infof("Adding additional staker to the network...")
	nonBootValidatorClient := allAvalancheClients[nonBootValidatorServiceID]
	highLevelExtraStakerClient := helpers.NewRPCWorkFlowRunner(
		nonBootValidatorClient,
		api.UserPass{Username: stakerUsername, Password: stakerPassword},
		networkAcceptanceTimeout)
	if _, err := highLevelExtraStakerClient.ImportGenesisFundsAndStartValidating(seedAmount, stakeAmount); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add extra staker."))
	}

	logrus.Infof("Sleeping %v seconds before verifying that the network has fully connected to the new staker...", test.FullyConnectedDelay.Seconds())
	// Give time for the new validator to propagate via gossip
	time.Sleep(70 * time.Second)

	logrus.Infof("Verifying that the network is fully connected...")
	stakerIDs[nonBootValidatorServiceID] = true
	/*
		After gossip, we expect the peers list to look like:
		1) No node has itself in its peers list
		2) The validators will have ALL other nodes in the network (propagated via gossip)
		3) The non-validators will have all the validators in the network (propagated via gossip)
	*/
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIDs, stakerIDs, allNodeIDs, allAvalancheClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying that the network is fully connected after gossip"))
	}
	logrus.Infof("The network is fully connected.")
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkFullyConnectedTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	serviceConfigs := map[networks.ConfigurationID]avalancheNetwork.TestAvalancheNetworkServiceConfig{
		normalNodeConfigID: *avalancheNetwork.NewTestAvalancheNetworkServiceConfig(
			true,
			avalancheService.DEBUG,
			test.ImageName,
			2,
			2,
			2*time.Second,
			make(map[string]string),
		),
	}
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{
		nonBootValidatorServiceID:    normalNodeConfigID,
		nonBootNonValidatorServiceID: normalNodeConfigID,
	}
	return avalancheNetwork.NewTestAvalancheNetworkLoader(
		true,
		test.ImageName,
		avalancheService.DEBUG,
		2,
		2,
		nil,
		0,
		2*time.Second,
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
This helper function will grab node IDs and Avalanche clients
*/
func getNodeIDsAndClients(
	testContext testsuite.TestContext,
	network avalancheNetwork.TestAvalancheNetwork,
	allServiceIDs map[networks.ServiceID]bool,
) (allNodeIDs map[networks.ServiceID]string, allAvalancheClients map[networks.ServiceID]*avalancheService.Client) {
	allAvalancheClients = make(map[networks.ServiceID]*avalancheService.Client)
	allNodeIDs = make(map[networks.ServiceID]string)
	for serviceID := range allServiceIDs {
		client, err := network.GetAvalancheClient(serviceID)
		if err != nil {
			testContext.Fatal(stacktrace.Propagate(err, "An error occurred getting the Avalanche client for service with ID %v", serviceID))
		}
		allAvalancheClients[serviceID] = client
		nodeID, err := client.InfoAPI().GetNodeID()

		if err != nil {
			testContext.Fatal(stacktrace.Propagate(err, "An error occurred getting the Avalanche node ID for service with ID %v", serviceID))
		}
		allNodeIDs[serviceID] = nodeID
	}
	return
}
