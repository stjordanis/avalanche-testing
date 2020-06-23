package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

type FiveNodeStakingNetworkFullyConnectedTest struct{}
func (test FiveNodeStakingNetworkFullyConnectedTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.NNodeGeckoNetwork)
	networkIdSet := map[string]bool{}
	numNodes := castedNetwork.GetNumberOfNodes()
	referenceNodeClient, err := castedNetwork.GetGeckoClient(0)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get reference client"))
	}
	testAccountAddress, err := referenceNodeClient.PChainApi().CreateAccount(
		"test",
		"test34test!23",
		"24jUJ9vZexUM6expyMcT48LBx27k1m7xpraoV62oSQAHdziao5")
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to create a test account on the network."))
	}
	logrus.Debugf("Test account address: %s", testAccountAddress)

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
		txnId, err := referenceNodeClient.PChainApi().AddDefaultSubnetValidator(id, 0, 9999999999, 1, 1, "", 1)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not add subnet validator %s", id))
		}
		logrus.Debugf("Transaction for adding subnet validator %s: %s", id, txnId)
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
		peerSet := map[string]bool{}
		for _, peer := range peers {
			peerSet[peer.Id] = true
			// verify that peer is inside the networkIdSet
			// context.AssertTrue(networkIdSet[peer.Id])
		}
		logrus.Debugf("Peers for node %d are %+v", i, peerSet)
		// verify that every other peer (besides the node itself) is represented in the peer list.
		// context.AssertTrue(len(peerSet) == numNodes - 1)
	}
}

func (s FiveNodeStakingNetworkFullyConnectedTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return ava_networks.NewNNodeGeckoNetworkLoader(5, 1, true)
}

func (test FiveNodeStakingNetworkFullyConnectedTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

type FiveNodeStakingNetworkBasicTest struct{}
func (test FiveNodeStakingNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.NNodeGeckoNetwork)

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

func (s FiveNodeStakingNetworkBasicTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return ava_networks.NewNNodeGeckoNetworkLoader(5, 1, true)
}

func (test FiveNodeStakingNetworkBasicTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

// =============== Get Validators Test ==================================
type FiveNodeStakingNetworkGetValidatorsTest struct{}
func (test FiveNodeStakingNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.NNodeGeckoNetwork)

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
	return ava_networks.NewNNodeGeckoNetworkLoader(5, 1, true)
}

func (test FiveNodeStakingNetworkGetValidatorsTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

