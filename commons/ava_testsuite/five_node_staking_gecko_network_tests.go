package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks/fixed_gecko_network"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	USERNAME = "test"
	PASSWORD = "test34test!23"
	PRIVATE_KEY = "24jUJ9vZexUM6expyMcT48LBx27k1m7xpraoV62oSQAHdziao5"
	// defined in Gecko codebase for default genesis block
	// PREFUNDED_ADDRESS = "6Y3kysjF9jnHnYkdS9yGAuoHyae2eNmeV"
	PREFUNDED_ADDRESS_PRIVATE_KEY = "ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
)

func addNodeAsValidator(client *gecko_client.GeckoClient) (string, error) {
	_, err := client.KeystoreApi().CreateUser(USERNAME, PASSWORD)
	if err != nil {
		stacktrace.Propagate(err, "Could not create user.")
	}
	nodeId, err := client.AdminApi().GetNodeId()
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not get node id")
	}
	genesisAccountAddress, err := client.XChainApi().ImportKey(
		USERNAME,
		PASSWORD,
		PREFUNDED_ADDRESS_PRIVATE_KEY)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to take control of genesis account.")
	}
	logrus.Debugf("Adding Node %s as a validator.", nodeId)
	logrus.Debugf("Genesis Address: %s.", genesisAccountAddress)
	testAccountAddress, err := client.PChainApi().CreateAccount(USERNAME, PASSWORD, nil)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create account on PChain.")
	}
	logrus.Debugf("Test account address: %s", testAccountAddress)
	unsignedTxnId, err := client.XChainApi().ExportAVA(testAccountAddress, 500, USERNAME, PASSWORD)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to export AVA.")
	}
	txnStatus := ""
	tries := 0
	for txnStatus != gecko_client.TXN_ACCEPTED && tries < 10 {
		time.Sleep(1*time.Second)
		tries++
		txnStatus, err := client.XChainApi().GetTxStatus(unsignedTxnId)
		if err != nil {
			return "", stacktrace.Propagate(err,"Failed to get transaction status for %s", unsignedTxnId)
		}
		logrus.Debugf("Export AVA from XChain: Transaction %s , Status: %s", unsignedTxnId, txnStatus)
	}
	genesisBalance, err := client.XChainApi().GetBalance(genesisAccountAddress, "AVA")
	logrus.Debugf("Genesis Account Balance: %+v", genesisBalance)
	unsignedTxnId, err = client.PChainApi().ImportAVA(USERNAME, PASSWORD, testAccountAddress, 1)
	//testAccountAddress = strings.TrimPrefix(testAccountAddress, "X-")
	unsignedTxnId, err = client.PChainApi().AddDefaultSubnetValidator(
		nodeId,
		time.Now().Add(5 * time.Hour).Unix(),
		time.Now().Add(50 * time.Hour).Unix(),
		100000,
		1,
		testAccountAddress,
		1)
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not add subnet validator %s", nodeId)
	}
	signedTxnId, err := client.PChainApi().Sign(unsignedTxnId, testAccountAddress, USERNAME, PASSWORD)
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not sign transaction to add validator %s", nodeId)
	}
	txnId, err := client.PChainApi().IssueTx(signedTxnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not issue txn to add validator %s", nodeId)
	}
	logrus.Debugf("Transaction for adding subnet validator %s: %s", nodeId, txnId)
	return nodeId, nil
}

type FiveNodeStakingNetworkFullyConnectedTest struct{}
func (test FiveNodeStakingNetworkFullyConnectedTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(fixed_gecko_network.FixedGeckoNetwork)
	networkIdSet := map[string]bool{}
	numNodes := castedNetwork.GetNumberOfNodes()
	referenceNodeClient, err := castedNetwork.GetGeckoClient(0)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get reference client"))
	}
    _, err = referenceNodeClient.KeystoreApi().CreateUser(USERNAME, PASSWORD)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not create user."))
	}

	// collect set of IDs in network
	for i := 0; i < numNodes; i++ {
		client, err := castedNetwork.GetGeckoClient(i)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client"))
		}
		validators, err := client.PChainApi().GetCurrentValidators(nil)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get current validators."))
		}
		logrus.Debugf("Current validators: %+v", validators)
		peers, err := client.AdminApi().GetPeers()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get current peers."))
		}
		logrus.Debugf("Current peers: %+v", peers)
		id, err := addNodeAsValidator(client)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not add node as validator."))
		}
		networkIdSet[id] = true
	}
	logrus.Debugf("Network ID Set: %+v", networkIdSet)
	// wait for all nodes to become validators by their timestamp
	// time.Sleep(time.Second * 10)
	// verify peer lists have set of IDs in network, except their own
	for i := 0; i < numNodes; i++ {
		client, err := castedNetwork.GetGeckoClient(i)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client"))
		}
		validators, err := client.PChainApi().GetPendingValidators(nil)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get peers"))
		}
		validatorSet := map[string]bool{}
		for _, validator := range validators {
			validatorSet[validator.Id] = true
			// verify that peer is inside the networkIdSet
			// context.AssertTrue(networkIdSet[peer.Id])
		}
		logrus.Debugf("Validators for node %d are %+v", i, validatorSet)
		// verify that every other peer (besides the node itself) is represented in the peer list.
		// context.AssertTrue(len(validatorSet) == numNodes - 1)
	}
}

func (test FiveNodeStakingNetworkFullyConnectedTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(5, 1, true)
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
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(5, 1, true)
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
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(5, 1, true)
}

func (test FiveNodeStakingNetworkGetValidatorsTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

