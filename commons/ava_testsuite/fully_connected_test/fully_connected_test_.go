package fully_connected_test

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite/rpc_workflow_runner"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite/verifier"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"time"
)

const (
	stakerUsername           = "staker"
	stakerPassword           = "test34test!23"
	seedAmount               = int64(50000000000000)
	stakeAmount              = int64(30000000000000)

	normalNodeConfigId networks.ConfigurationID = 0

	networkAcceptanceTimeoutRatio = 0.3
	nonBootValidatorServiceId networks.ServiceID = 0
	nonBootNonValidatorServiceId networks.ServiceID = 1
)

type StakingNetworkFullyConnectedTest struct{
	ImageName string
	Verifier  verifier.NetworkStateVerifier
}
func (test StakingNetworkFullyConnectedTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))


	stakerIds := castedNetwork.GetAllBootServiceIds()
	allServiceIds := make(map[networks.ServiceID]bool)
	for stakerId, _ := range stakerIds {
		allServiceIds[stakerId] = true
	}
	// Add our custom nodes
	allServiceIds[nonBootValidatorServiceId] = true
	allServiceIds[nonBootNonValidatorServiceId] = true

	allNodeIds, allGeckoClients := getNodeIdsAndClients(context, castedNetwork, allServiceIds)
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIds, stakerIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}

	nonBootValidatorClient := allGeckoClients[nonBootValidatorServiceId]
	highLevelExtraStakerClient := rpc_workflow_runner.NewRpcWorkflowRunner(
		nonBootValidatorClient,
		stakerUsername,
		stakerPassword,
		networkAcceptanceTimeout)
	if err := highLevelExtraStakerClient.GetFundsAndStartValidating(seedAmount, stakeAmount); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add extra staker."))
	}

	// Give time for the new validator to propagate via gossip
	time.Sleep(70 * time.Second)

	stakerIds[nonBootValidatorServiceId] = true

	/*
		After gossip, we expect the peers list to look like:
		1) No node has itself in its peers list
		2) The validators will have ALL other nodes in the network (propagated via gossip)
		3) The non-validators will have all the validators in the network (propagated via gossip)
	*/
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIds, stakerIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying that the network is fully connected after gossip"))
	}
}

func (test StakingNetworkFullyConnectedTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	serviceConfigs := map[networks.ConfigurationID]ava_networks.TestGeckoNetworkServiceConfig{
		normalNodeConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, test.ImageName, 2, 2),
	}
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{
		nonBootValidatorServiceId: normalNodeConfigId,
		nonBootNonValidatorServiceId: normalNodeConfigId,
	}
	return ava_networks.NewTestGeckoNetworkLoader(
		true,
		test.ImageName,
		ava_services.LOG_LEVEL_DEBUG,
		2,
		2,
		serviceConfigs,
		desiredServices)
}

func (test StakingNetworkFullyConnectedTest) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

func (test StakingNetworkFullyConnectedTest) GetSetupBuffer() time.Duration {
	// TODO drop this when the availabilityChecker doesn't have a sleep (because we spin up a bunch of nodes before running the test)
	return 6 * time.Minute
}

// ================ Helper functions =========================
/*
This helper function will grab node IDs and Gecko clients
*/
func getNodeIdsAndClients(
	testContext testsuite.TestContext,
	network ava_networks.TestGeckoNetwork,
	allServiceIds map[networks.ServiceID]bool,
) (allNodeIds map[networks.ServiceID]string, allGeckoClients map[networks.ServiceID]*gecko_client.GeckoClient){
	allGeckoClients = make(map[networks.ServiceID]*gecko_client.GeckoClient)
	allNodeIds = make(map[networks.ServiceID]string)
	for serviceId, _ := range allServiceIds {
		client, err := network.GetGeckoClient(serviceId)
		if err != nil {
			testContext.Fatal(stacktrace.Propagate(err, "An error occurred getting the Gecko client for service with ID %v", serviceId))
		}
		allGeckoClients[serviceId] = client
		nodeId, err := client.InfoApi().GetNodeId()
		if err != nil {
			testContext.Fatal(stacktrace.Propagate(err, "An error occurred getting the Gecko node ID for service with ID %v", serviceId))
		}
		allNodeIds[serviceId] = nodeId
	}
	return
}
