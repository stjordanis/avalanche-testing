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
	// defined in Gecko codebase for default genesis block
	// PREFUNDED_ADDRESS = "6Y3kysjF9jnHnYkdS9yGAuoHyae2eNmeV"
)


type FiveNodeStakingNetworkXChainTransferTest struct{}
func (test FiveNodeStakingNetworkXChainTransferTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(fixed_gecko_network.FixedGeckoNetwork)
	testAmount := 10000
	referenceNodeClient, err := castedNetwork.GetGeckoClient(4)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get reference client"))
	}
	rpcManager := RpcManager{
		client: referenceNodeClient,
		testNet: ava_default_testnet.DefaultTestNet,
	}
	address, err := rpcManager.createAndSeedXChainAccountFromGenesis(USERNAME, PASSWORD, testAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not seed XChain account from Genesis."))
	}
	balance, err := referenceNodeClient.XChainApi().GetBalance(address, "AVA")
	context.AssertTrue(balance.Balance == strconv.Itoa(testAmount))
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


func createAndSeedXChainAccountFromGenesis(client *gecko_client.GeckoClient, username string, password string, amount int) (string, error) {
	time.Sleep(time.Second * 30)
	_, err := client.KeystoreApi().CreateUser(username, password)
	if err != nil {
		stacktrace.Propagate(err, "Could not create user.")
	}
	_, err = client.KeystoreApi().CreateUser(GENESIS_USERNAME, GENESIS_PASSWORD)
	if err != nil {
		stacktrace.Propagate(err, "Could not create genesis user.")
	}
	nodeId, err := client.AdminApi().GetNodeId()
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not get node id")
	}
	genesisAccountAddress, err := client.XChainApi().ImportKey(
		GENESIS_USERNAME,
		GENESIS_PASSWORD,
		ava_default_testnet.DefaultTestNet.FundedAddresses.PrivateKey)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to take control of genesis account.")
	}
	logrus.Debugf("Adding Node %s as a validator.", nodeId)
	logrus.Debugf("Genesis Address: %s.", genesisAccountAddress)
	testAccountAddress, err := client.XChainApi().CreateAddress(username, password)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create address on XChain.")
	}
	logrus.Debugf("Test account address: %s", testAccountAddress)
	txnId, err := client.XChainApi().Send(amount, "AVA", testAccountAddress, GENESIS_USERNAME, GENESIS_PASSWORD)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to send AVA to test account address %s", testAccountAddress)
	}
	status := ""
	for status != ACCEPTED_STATUS {
		status, err = client.XChainApi().GetTxStatus(txnId)
		if err != nil {
			return "", stacktrace.Propagate(err,"Failed to get status.")
		}
		time.Sleep(time.Second)
	}
	logrus.Debugf("Transaction status for send transaction: %s", status)
	return testAccountAddress, nil
}


