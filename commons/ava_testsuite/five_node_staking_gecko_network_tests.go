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
	"strconv"
	"time"
)

const (
	USERNAME = "test"
	PASSWORD = "test34test!23"
	SEED_AMOUNT = 1000000

	NODE_SERVICE_ID       = 0
	NORMAL_NODE_CONFIG_ID = 0

	// The configuration ID of a service
	SAME_CERT_CONIFG_ID = 1

)

type FiveNodeStakingNetworkRpcWorkflowTest struct{}
func (test FiveNodeStakingNetworkRpcWorkflowTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	referenceNodeClient, err := castedNetwork.GetGeckoClient(NODE_SERVICE_ID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get reference client"))
	}
	rpcManager := ava_networks.NewHighLevelGeckoClient(
		referenceNodeClient,
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
	// TODO TODO TODO Test adding stakers
	// TODO TODO TODO Test adding delegators
	// TODO TODO TODO Test transferring staking rewards back to XChain
}
func (test FiveNodeStakingNetworkRpcWorkflowTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getFiveNodeStakingLoader()
}
func (test FiveNodeStakingNetworkRpcWorkflowTest) GetTimeout() time.Duration {
	return 60 * time.Second
}


type FiveNodeStakingNetworkFullyConnectedTest struct{}
func (test FiveNodeStakingNetworkFullyConnectedTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	allServiceIds := castedNetwork.GetAllBootServiceIds()
	allServiceIds[NODE_SERVICE_ID] = true

	// collect set of IDs in network
	nodeIdSet := map[string]bool{}
	for serviceId, _ := range allServiceIds {
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
	for serviceId, _ := range allServiceIds {
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

// =============== Duplicate Node ID Test ==============================
type FiveNodeStakingNetworkDuplicateIdTest struct {}
func (f FiveNodeStakingNetworkDuplicateIdTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	allServiceIds := castedNetwork.GetAllBootServiceIds()
	allServiceIds[NODE_SERVICE_ID] = true

	originalServiceIds := make(map[int]bool)
	for serviceId, _ := range allServiceIds {
		originalServiceIds[serviceId] = true
	}

	// Verify that everybody has everyone else as peers before we add the services with the duplicate nodes
	allGeckoNodeIds := make(map[string]bool)
	allGeckoClients := make(map[int]*gecko_client.GeckoClient)
	for serviceId, _ := range allServiceIds {
		geckoClient, err := castedNetwork.GetGeckoClient(serviceId)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Error getting Gecko client for service with ID %v", serviceId))
		}
		allGeckoClients[serviceId] = geckoClient

		adminApi := geckoClient.AdminApi()
		peers, err := adminApi.GetPeers()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get peers from service with ID %v", serviceId))
		}
		context.AssertTrue(len(peers) == len(allServiceIds) - 1)
		for _, peer := range peers {
			allGeckoNodeIds[peer.Id] = true
		}
	}

	// We'll need these later
	originalGeckoNodeIds := make(map[string]bool)
	for nodeId, _ := range allGeckoNodeIds {
		originalGeckoNodeIds[nodeId] = true
	}

	// Add the first dupe node ID (should look normal from a network perspective
	badServiceId1 := 1
	checker1, err := castedNetwork.AddService(SAME_CERT_CONIFG_ID, badServiceId1)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to create first dupe node ID service with ID %v", badServiceId1))
	}
	if err := checker1.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred waiting for first dupe node ID service to start"))
	}
	allServiceIds[badServiceId1] = true

	badServiceClient1, err := castedNetwork.GetGeckoClient(badServiceId1)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred getting the Gecko client for the first dupe node ID service with ID %v", badServiceId1))
	}
	allGeckoClients[badServiceId1] = badServiceClient1

	badServiceNodeId1, err := badServiceClient1.AdminApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get node ID from first dupe node ID service with ID %v", badServiceId1))
	}
	allGeckoNodeIds[badServiceNodeId1] = true

	// Verify that the new node got accepted by everyone
	for serviceId, _ := range allServiceIds {
		client := allGeckoClients[serviceId]

		adminApi := client.AdminApi()
		peers, err := adminApi.GetPeers()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get peers from service with ID %v", serviceId))
		}
		context.AssertTrue(len(peers) == len(allServiceIds) - 1)
		for _, peer := range peers {
			_, found := allGeckoNodeIds[peer.Id]
			context.AssertTrue(found)
		}
	}

	// Now, add a second node with the same ID
	badServiceId2 := 1
	checker2, err := castedNetwork.AddService(SAME_CERT_CONIFG_ID, badServiceId2)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to create second dupe node ID service with ID %v", badServiceId2))
	}
	if err := checker2.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred waiting for second dupe node ID service to start"))
	}
	allServiceIds[badServiceId2] = true

	badServiceClient2, err := castedNetwork.GetGeckoClient(badServiceId2)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred getting the Gecko client for the second dupe node ID service with ID %v", badServiceId2))
	}
	allGeckoClients[badServiceId2] = badServiceClient2

	badServiceNodeId2, err := badServiceClient2.AdminApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get node ID from first dupe node ID service with ID %v", badServiceId2))
	}
	allGeckoNodeIds[badServiceNodeId2] = true

	// At this point, it's undefined what happens with the two nodes with duplicate IDs; verify that the original nodes
	//  network operates normally amongst themselves
	for serviceId, _ := range originalServiceIds {
		client := allGeckoClients[serviceId]
		peers, err := client.AdminApi().GetPeers()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "An error occurred getting peers from service with ID %v", serviceId))
		}

		// Verify we have, at minimum, all the original nodes
		originalGeckoNodeIdsSeen := make(map[string]bool)
		for _, peer := range peers {
			peerId := peer.Id
			if _, found := originalGeckoNodeIds[peerId]; found {
				originalGeckoNodeIdsSeen[peerId] = true
			}
		}
		context.AssertTrue(len(originalGeckoNodeIdsSeen) == len(originalGeckoNodeIds) - 1)
	}

	// Now, kill the first dupe node to leave only the second (who everyone should connect with)
	if err := castedNetwork.RemoveService(badServiceId1); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not remove the first service with duped node ID"))
	}
	delete(allServiceIds, badServiceId1)
	delete(allGeckoClients, badServiceId1)
	delete(allGeckoNodeIds, badServiceNodeId1)

	// Now that the first duped node ID is gone, leaving only the second, verify everyone's happy again
	for serviceId, _ := range allServiceIds {
		client := allGeckoClients[serviceId]

		adminApi := client.AdminApi()
		peers, err := adminApi.GetPeers()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get peers from service with ID %v", serviceId))
		}
		context.AssertTrue(len(peers) == len(allServiceIds) - 1)
		for _, peer := range peers {
			_, found := allGeckoNodeIds[peer.Id]
			context.AssertTrue(found)
		}
	}
}

func (f FiveNodeStakingNetworkDuplicateIdTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getFiveNodeStakingLoader()
}

func (f FiveNodeStakingNetworkDuplicateIdTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

// =============== Helper functions =============================

func getFiveNodeStakingLoader() (testsuite.TestNetworkLoader, error) {
	serviceConfigs := map[int]ava_networks.TestGeckoNetworkServiceConfig{
		NORMAL_NODE_CONFIG_ID: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG),
		SAME_CERT_CONIFG_ID: *ava_networks.NewTestGeckoNetworkServiceConfig(false, ava_services.LOG_LEVEL_DEBUG),
	}
	return ava_networks.NewTestGeckoNetworkLoader(
		ava_services.LOG_LEVEL_DEBUG,
		true,
		serviceConfigs,
		map[int]int{
			NODE_SERVICE_ID: NORMAL_NODE_CONFIG_ID,
		},
		2,
		2)
}
