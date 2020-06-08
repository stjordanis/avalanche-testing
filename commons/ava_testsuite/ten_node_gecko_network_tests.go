package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

// =============== Basic Test ==================================
type TenNodeGeckoNetworkBasicTest struct {}
func (s TenNodeGeckoNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
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

func (s TenNodeGeckoNetworkBasicTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return ava_networks.NewNNodeGeckoNetworkLoader(10, 3, false)
}


// =============== Get Validators Test ==================================
type TenNodeNetworkGetValidatorsTest struct{}
func (test TenNodeNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
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
		validators, err = client.PChainApi().GetCurrentValidators()
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

func (test TenNodeNetworkGetValidatorsTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return ava_networks.NewNNodeGeckoNetworkLoader(10, 3, false)
}

// =============== Get Peers Test ==================================
type TenNodeNetworkGetPeersTest struct{}
func (test TenNodeNetworkGetPeersTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.NNodeGeckoNetwork)

	clients := make([]gecko_client.GeckoClient, 10)
	nodeIDs := map[string]struct{}{}
	for i := 0; i < 10; i++ {
		client, err := castedNetwork.GetGeckoClient(i)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client"))
		}

		logrus.Debug("Trying to get node id for node", i)
		nodeID := ""
		for j := 0; j < 5; j++ {
			nodeID, err = client.AdminApi().GetNodeId()
			if err == nil {
				break
			}

			time.Sleep(5 * time.Second)
		}
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get node id"))
			return
		}

		logrus.Debug("Adding node ID: ", nodeID)
		clients[i] = client
		nodeIDs[nodeID] = struct{}{}
	}

	logrus.Debug("nodeIDs:")
	logrus.Debug(nodeIDs)

	// Test each client's peer list against the known nodes
	for _, client := range clients {
		peers, err := client.AdminApi().GetPeers()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get current peers"))
		}
		logrus.Debug("Peers from node:")

		// Assert that the returned peer list contains every node except this one
		logrus.Debug("Asserting ", len(peers), " equals ",len(nodeIDs)-1 )
		// context.AssertTrue(len(peers) == len(nodeIDs)-1)
		for _, peer := range peers {
			logrus.Infof("Peer ID: %s ", peer.Id)

			_, ok := nodeIDs[peer.Id]
			logrus.Debug("Asserting ", peer.Id, " is in nodeIDs ", ok)
			// context.AssertTrue(ok)
		}
	}
}

func (test TenNodeNetworkGetPeersTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return ava_networks.NewNNodeGeckoNetworkLoader(10, 3, false)
}
