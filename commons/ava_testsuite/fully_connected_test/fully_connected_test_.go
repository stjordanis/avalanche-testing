package fully_connected_test

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
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

	normalNodeConfigId = 0

	nonBootValidatorServiceId = 0
	nonBootNonValidatorServiceId = 1
)

type StakingNetworkFullyConnectedTest struct{
	ImageName string
	Verifier  verifier.NetworkStateVerifier
}
func (test StakingNetworkFullyConnectedTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	nonBootValidatorServiceId := 0
	nonBootNonValidatorServiceId := 1

	stakerIds := castedNetwork.GetAllBootServiceIds()
	allServiceIds := make(map[int]bool)
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
	highLevelExtraStakerClient := ava_networks.NewHighLevelGeckoClient(
		nonBootValidatorClient,
		stakerUsername,
		stakerPassword)
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
	serviceConfigs := map[int]ava_networks.TestGeckoNetworkServiceConfig{
		normalNodeConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, test.ImageName, 2, 2),
	}
	desiredServices := map[int]int{
		nonBootValidatorServiceId: normalNodeConfigId,
		nonBootNonValidatorServiceId: normalNodeConfigId,
	}
	return ava_networks.NewTestGeckoNetworkLoader(
		ava_services.LOG_LEVEL_DEBUG,
		true,
		serviceConfigs,
		desiredServices,
		2,
		2)
}

func (test StakingNetworkFullyConnectedTest) GetTimeout() time.Duration {
	return 120 * time.Second
}

// ================ Helper functions =========================
/*
This helper function will grab node IDs and Gecko clients
*/
func getNodeIdsAndClients(
	testContext testsuite.TestContext,
	network ava_networks.TestGeckoNetwork,
	allServiceIds map[int]bool) (allNodeIds map[int]string, allGeckoClients map[int]*gecko_client.GeckoClient){
	allGeckoClients = make(map[int]*gecko_client.GeckoClient)
	allNodeIds = make(map[int]string)
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
