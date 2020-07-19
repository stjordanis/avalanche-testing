package fully_connected_test

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"time"
)

type StakingNetworkFullyConnectedTest struct{
	imageName string
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
	if err := verifyNetworkFullyConnected(allServiceIds, stakerIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}

	nonBootValidatorClient := allGeckoClients[nonBootValidatorServiceId]
	highLevelExtraStakerClient := ava_networks.NewHighLevelGeckoClient(
		nonBootValidatorClient,
		STAKER_USERNAME,
		STAKER_PASSWORD)
	if err := highLevelExtraStakerClient.GetFundsAndStartValidating(SEED_AMOUNT, STAKE_AMOUNT); err != nil {
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
	if err := verifyNetworkFullyConnected(allServiceIds, stakerIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying that the network is fully connected after gossip"))
	}
}

func (test StakingNetworkFullyConnectedTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	return getStakingNetworkLoader(map[int]int{
		// TODO Once this is split into its own file, make these constants
		0: NORMAL_NODE_CONFIG_ID,
		1: NORMAL_NODE_CONFIG_ID,
	}, test.imageName)
}

func (test StakingNetworkFullyConnectedTest) GetTimeout() time.Duration {
	return 120 * time.Second
}
