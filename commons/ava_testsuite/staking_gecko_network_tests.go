package ava_testsuite

// TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO
//  Rename this entire file and everything in it to emphasize the "staking" aspect, not the number of nodes (because the
//  number of nodes doesn't really matter)
// TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	STAKER_USERNAME           = "staker"
	STAKER_PASSWORD           = "test34test!23"
	DELEGATOR_USERNAME           = "delegator"
	DELEGATOR_PASSWORD           = "test34test!23"
	SEED_AMOUNT               = int64(50000000000000)
	STAKE_AMOUNT              = int64(30000000000000)
	DELEGATOR_AMOUNT              = int64(30000000000000)
	NODE_SERVICE_ID           = 0
	NODE_CONFIG_ID            = 0
	DELEGATOR_NODE_SERVICE_ID = 1
	DELEGATOR_NODE_CONFIG_ID  = 1
)

type StakingNetworkRpcWorkflowTest struct{}
func (test StakingNetworkRpcWorkflowTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	stakerClient, err := castedNetwork.GetGeckoClient(NODE_SERVICE_ID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get staker client"))
	}
	delegatorClient, err := castedNetwork.GetGeckoClient(DELEGATOR_NODE_SERVICE_ID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get delegator client"))
	}
	stakerNodeId, err := stakerClient.AdminApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get staker node ID."))
	}
	delegatorNodeId, err := stakerClient.AdminApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get delegator node ID."))
	}
	highLevelStakerClient := ava_networks.NewHighLevelGeckoClient(
		stakerClient,
		STAKER_USERNAME,
		STAKER_PASSWORD)
	highLevelDelegatorClient := ava_networks.NewHighLevelGeckoClient(
		delegatorClient,
		DELEGATOR_USERNAME,
		DELEGATOR_PASSWORD)
	stakerXchainAddress, err := highLevelStakerClient.CreateAndSeedXChainAccountFromGenesis(SEED_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not seed XChain account from Genesis."))
	}
	stakerPchainAddress, err := highLevelStakerClient.TransferAvaXChainToPChain(SEED_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information"))
	}
	_, err = highLevelDelegatorClient.CreateAndSeedXChainAccountFromGenesis(SEED_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not seed XChain account from Genesis."))
	}
	delegatorPchainAddress, err := highLevelDelegatorClient.TransferAvaXChainToPChain(SEED_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information"))
	}
	// Adding stakers
	err = highLevelStakerClient.AddValidatorOnSubnet(stakerNodeId, stakerPchainAddress, STAKE_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not add staker %s to default subnet.", stakerNodeId))
	}
	currentStakers, err := stakerClient.PChainApi().GetCurrentValidators(nil)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
	}
	logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	context.AssertTrue(len(currentStakers) == 6)
	// Adding delegators
	err = highLevelDelegatorClient.AddDelegatorOnSubnet(stakerNodeId, delegatorPchainAddress, DELEGATOR_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not add delegator %s to default subnet.", delegatorNodeId))
	}
	/*
		Currently no way to verify rewards for stakers and delegators because rewards are
		only paid out at the end of the staking period, and the staking period must last at least
		24 hours. This is far too long to be able to test in a CI scenario.
	 */
	remainingStakerAva := SEED_AMOUNT - STAKE_AMOUNT
	highLevelStakerClient.TransferAvaPChainToXChain(stakerPchainAddress, stakerXchainAddress, remainingStakerAva)
	xchainAccountInfo, err := stakerClient.XChainApi().GetBalance(stakerXchainAddress, ava_networks.AVA_ASSET_ID)
	logrus.Debugf("XChain account info: %+v", xchainAccountInfo)
	context.AssertTrue(string(remainingStakerAva) == xchainAccountInfo.Balance)
}
func (test StakingNetworkRpcWorkflowTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getFiveNodeStakingLoader()
}
func (test StakingNetworkRpcWorkflowTest) GetTimeout() time.Duration {
	return 90 * time.Second
}


type FiveNodeStakingNetworkFullyConnectedTest struct{}
func (test FiveNodeStakingNetworkFullyConnectedTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	bootServiceIds := castedNetwork.GetAllBootServiceIds()
	allServiceIds := append(bootServiceIds, NODE_SERVICE_ID)

	// collect set of IDs in network
	nodeIdSet := map[string]bool{}
	for _, serviceId := range allServiceIds {
		client, err := castedNetwork.GetGeckoClient(serviceId)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client for service with ID %v", serviceId))
		}
		id, err := client.AdminApi().GetNodeId()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get node ID of service with ID %v", serviceId))
		}
		nodeIdSet[id] = true
	}

	logrus.Debugf("Network ID Set: %+v", nodeIdSet)

	// verify peer lists have set of IDs in network, except their own
	for _, serviceId := range allServiceIds {
		client, err := castedNetwork.GetGeckoClient(serviceId)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client for service with ID %v", serviceId))
		}
		peers, err := client.AdminApi().GetPeers()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get peers of service with ID %v", serviceId))
		}
		logrus.Debugf("Peer set: %+v", peers)

		peerSet := map[string]bool{}
		for _, peer := range peers {
			peerSet[peer.Id] = true
			// verify that peer is inside the nodeIdSet
			context.AssertTrue(nodeIdSet[peer.Id])
		}
		// verify that every other peer (besides the node itself) is represented in the peer list.
		context.AssertTrue(len(peerSet) == len(allServiceIds) - 1)
	}
}

func (test FiveNodeStakingNetworkFullyConnectedTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getFiveNodeStakingLoader()
}

func (test FiveNodeStakingNetworkFullyConnectedTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

type FiveNodeStakingNetworkBasicTest struct{}
func (test FiveNodeStakingNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	// TODO check ALL nodes!
	client, err := castedNetwork.GetGeckoClient(0)
	if err != nil {
	context.Fatal(stacktrace.Propagate(err, "Could not get client"))
	}

	peers, err := client.AdminApi().GetPeers()
	if err != nil {
	context.Fatal(stacktrace.Propagate(err, "Could not get peers"))
	}

	context.AssertTrue(len(peers) == 9)
}

func (test FiveNodeStakingNetworkBasicTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getFiveNodeStakingLoader()
}

func (test FiveNodeStakingNetworkBasicTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

// =============== Get Validators Test ==================================
type FiveNodeStakingNetworkGetValidatorsTest struct{}
func (test FiveNodeStakingNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	// TODO we need to make sure ALL the nodes agree about validators!
	client, err := castedNetwork.GetGeckoClient(0)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get client"))
	}

	// TODO This retry logic is only necessary because there's not a way for Ava nodes to reliably report
	//  bootstrapping as complete; remove it when Gecko can report successful bootstrapping
	var validators []gecko_client.Validator
	for i := 0; i < 5; i++ {
		validators, err = client.PChainApi().GetCurrentValidators(nil)
		if err == nil {
			break
		}
		logrus.Error(stacktrace.Propagate(err, "Could not get current validators; sleeping for 5 seconds..."))
		time.Sleep(5 * time.Second)
	}
	// TODO This should go away as soon as Ava can reliably report bootstrapping as complete
	if validators == nil {
		context.Fatal(stacktrace.NewError("Could not get validators even after retrying!"))
	}

	for _, validator := range validators {
		logrus.Infof("Validator ID: %s", validator.Id)
	}
	// TODO change this to be specific
	context.AssertTrue(len(validators) >= 1)
}

func (test FiveNodeStakingNetworkGetValidatorsTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getFiveNodeStakingLoader()
}

func (test FiveNodeStakingNetworkGetValidatorsTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

// =============== Helper functions =============================

// TODO TODO TODO Rename this
func getFiveNodeStakingLoader() (testsuite.TestNetworkLoader, error) {
	serviceConfigs := map[int]ava_networks.TestGeckoNetworkServiceConfig{
		NODE_CONFIG_ID:           *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG),
		DELEGATOR_NODE_CONFIG_ID: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG),
	}
	return ava_networks.NewTestGeckoNetworkLoader(
		ava_services.LOG_LEVEL_DEBUG,
		true,
		serviceConfigs,
		map[int]int{
			NODE_SERVICE_ID:           NODE_CONFIG_ID,
			DELEGATOR_NODE_SERVICE_ID: DELEGATOR_NODE_CONFIG_ID,
		},
		2,
		2)
}

