package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_default_testnet"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks/fixed_gecko_network"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	USERNAME = "test"
	PASSWORD = "test34test!23"
	SEED_AMOUNT = 1000000
	// Use 4 as a reference node for now, because it appears to handle bootstrapping more quickly than the first node.
	// If we use 0, we get intermittent test timeouts.
	// TODO TODO TODO When bootstrapping API is available, use that to make sure testnet is ready.
	REFERENCE_NODE_INDEX = 4
)

type FiveNodeStakingNetworkPChainImportTest struct{}
func (test FiveNodeStakingNetworkPChainImportTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(fixed_gecko_network.FixedGeckoNetwork)
	referenceNodeClient, err := castedNetwork.GetGeckoClient(REFERENCE_NODE_INDEX)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get reference client"))
	}
	rpcManager := NewRpcManager(
		referenceNodeClient,
		&ava_default_testnet.DefaultTestNet,
		USERNAME,
		PASSWORD)
	_, err = rpcManager.CreateAndSeedXChainAccountFromGenesis(SEED_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not seed XChain account from Genesis."))
	}
	pchainAddress, err := rpcManager.TransferAvaXChainToPChain(SEED_AMOUNT)
	pchainAccount, err := referenceNodeClient.PChainApi().GetAccount(pchainAddress)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get PChain account information"))
	}
	balance := pchainAccount.Balance
	context.AssertTrue(balance == strconv.Itoa(SEED_AMOUNT))
}
func (test FiveNodeStakingNetworkPChainImportTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(5, 5, true)
}
func (test FiveNodeStakingNetworkPChainImportTest) GetTimeout() time.Duration {
	return 60 * time.Second
}

type FiveNodeStakingNetworkXChainTransferTest struct{}
func (test FiveNodeStakingNetworkXChainTransferTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(fixed_gecko_network.FixedGeckoNetwork)
	referenceNodeClient, err := castedNetwork.GetGeckoClient(REFERENCE_NODE_INDEX)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get reference client"))
	}
	rpcManager := NewRpcManager(
		referenceNodeClient,
		&ava_default_testnet.DefaultTestNet,
		USERNAME,
		PASSWORD)
	address, err := rpcManager.CreateAndSeedXChainAccountFromGenesis(SEED_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not seed XChain account from Genesis."))
	}
	balance, err := referenceNodeClient.XChainApi().GetBalance(address, "AVA")
	context.AssertTrue(balance.Balance == strconv.Itoa(SEED_AMOUNT))
}
func (test FiveNodeStakingNetworkXChainTransferTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(5, 5, true)
}
func (test FiveNodeStakingNetworkXChainTransferTest) GetTimeout() time.Duration {
	return 60 * time.Second
}



type FiveNodeStakingNetworkFullyConnectedTest struct{}
func (test FiveNodeStakingNetworkFullyConnectedTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(fixed_gecko_network.FixedGeckoNetwork)
	networkIdSet := map[string]bool{}
	numNodes := castedNetwork.GetNumberOfNodes()

	// collect set of IDs in network
	for i := 0; i < numNodes; i++ {
		client, err := castedNetwork.GetGeckoClient(i)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client"))
		}
		id, err := client.AdminApi().GetNodeId()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client"))
		}
		networkIdSet[id] = true
	}
	logrus.Debugf("Network ID Set: %+v", networkIdSet)
	// verify peer lists have set of IDs in network, except their own
	for i := 0; i < numNodes; i++ {
		client, err := castedNetwork.GetGeckoClient(i)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client"))
		}
		peers, err := client.AdminApi().GetPeers()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get peers"))
		}
		logrus.Debugf("Peer set: %+v", peers)
		peerSet := map[string]bool{}
		for _, peer := range peers {
			peerSet[peer.Id] = true
			// verify that peer is inside the networkIdSet
			context.AssertTrue(networkIdSet[peer.Id])
		}
		// verify that every other peer (besides the node itself) is represented in the peer list.
		context.AssertTrue(len(peerSet) == numNodes - 1)
	}
}

func (test FiveNodeStakingNetworkFullyConnectedTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(5, 5, true)
}

func (test FiveNodeStakingNetworkFullyConnectedTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

type FiveNodeStakingNetworkBasicTest struct{}
func (test FiveNodeStakingNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(fixed_gecko_network.FixedGeckoNetwork)

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
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(5, 5, true)
}

func (test FiveNodeStakingNetworkBasicTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

// =============== Get Validators Test ==================================
type FiveNodeStakingNetworkGetValidatorsTest struct{}
func (test FiveNodeStakingNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(fixed_gecko_network.FixedGeckoNetwork)

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
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(5, 5, true)
}

func (test FiveNodeStakingNetworkGetValidatorsTest) GetTimeout() time.Duration {
	return 30 * time.Second
}
