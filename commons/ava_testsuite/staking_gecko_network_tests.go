package ava_testsuite
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
	STAKER_USERNAME           = "staker"
	STAKER_PASSWORD           = "test34test!23"
	DELEGATOR_USERNAME           = "delegator"
	DELEGATOR_PASSWORD           = "test34test!23"
	SEED_AMOUNT               = int64(50000000000000)
	STAKE_AMOUNT              = int64(30000000000000)
	DELEGATOR_AMOUNT              = int64(30000000000000)
	NODE_SERVICE_ID           = 0
	DELEGATOR_NODE_SERVICE_ID = 1

	NORMAL_NODE_CONFIG_ID = 0

	// The configuration ID of a service where all servies made with this configuration will have the same cert
	SAME_CERT_CONFIG_ID = 1

	validatorPropagationSleepTime = 5 * time.Second
)

// ================ RPC Workflow Test ===================================
type StakingNetworkRpcWorkflowTest struct {
	imageName string
}
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
	stakerNodeId, err := stakerClient.InfoApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get staker node ID."))
	}
	delegatorNodeId, err := delegatorClient.InfoApi().GetNodeId()
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
	// These networks take a little time to propagate current validators
	// TODO TODO TODO Implement retries against current validators endpoint until latest added is included
	time.Sleep(validatorPropagationSleepTime)
	currentStakers, err := stakerClient.PChainApi().GetCurrentValidators(nil)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
	}
	logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	actualNumStakers := len(currentStakers)
	expectedNumStakers := 6
	context.AssertTrue(actualNumStakers == expectedNumStakers, stacktrace.NewError("Actual number of stakers, %v, != expected number of stakers, %v", actualNumStakers, expectedNumStakers))
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
	_, err = highLevelStakerClient.TransferAvaPChainToXChain(stakerPchainAddress, stakerXchainAddress, remainingStakerAva)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to transfer Ava from PChain to XChain."))
	}
	xchainAccountInfo, err := stakerClient.XChainApi().GetBalance(stakerXchainAddress, ava_networks.AVA_ASSET_ID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get account info for account %v.", stakerXchainAddress))
	}
	actualRemainingAva := xchainAccountInfo.Balance
	expectedRemainingAva := strconv.FormatInt(remainingStakerAva, 10)
	context.AssertTrue(actualRemainingAva == expectedRemainingAva, stacktrace.NewError("Actual remaining Ava, %v, != expected remaining Ava, %v", actualRemainingAva, expectedRemainingAva))
}
func (test StakingNetworkRpcWorkflowTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getStakingNetworkLoader(map[int]int{
		NODE_SERVICE_ID:           NORMAL_NODE_CONFIG_ID,
		DELEGATOR_NODE_SERVICE_ID: NORMAL_NODE_CONFIG_ID,
	}, test.imageName)
}
func (test StakingNetworkRpcWorkflowTest) GetTimeout() time.Duration {
	return 90 * time.Second
}


// =================== Fully Connected Test ==============================
type StakingNetworkFullyConnectedTest struct{
	imageName string
}
func (test StakingNetworkFullyConnectedTest) Run(network interface{}, context testsuite.TestContext) {
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

func (test StakingNetworkFullyConnectedTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getStakingNetworkLoader(map[int]int{
		// TODO Once this is split into its own file, make these constants
		0: NORMAL_NODE_CONFIG_ID,
		1: NORMAL_NODE_CONFIG_ID,
	}, test.imageName)
}

func (test StakingNetworkFullyConnectedTest) GetTimeout() time.Duration {
	return 120 * time.Second
}

// =============== Duplicate Node ID Test ==============================
type StakingNetworkDuplicateNodeIdTest struct {
	imageName string
}
func (f StakingNetworkDuplicateNodeIdTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	bootServiceIds := castedNetwork.GetAllBootServiceIds()

	allServiceIds := make(map[int]bool)
	for bootServiceId, _ := range bootServiceIds {
		allServiceIds[bootServiceId] = true
	}
	allServiceIds[NODE_SERVICE_ID] = true

	allNodeIds, allGeckoClients := getNodeIdsAndClients(context, castedNetwork, allServiceIds)
	if err := verifyNetworkFullyConnected(allServiceIds, bootServiceIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}

	// We'll need these later
	originalServiceIds := make(map[int]bool)
	for serviceId, _ := range allServiceIds {
		originalServiceIds[serviceId] = true
	}

	logrus.Debugf("Service IDs before adding any nodes: %v", allServiceIds)
	logrus.Debugf("Gecko node IDs before adding any nodes: %v", allNodeIds)

	// Add the first dupe node ID (should look normal from a network perspective
	badServiceId1 := 1
	logrus.Info("Adding first node with soon-to-be-duplicated node ID...")
	checker1, err := castedNetwork.AddService(SAME_CERT_CONFIG_ID, badServiceId1)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to create first dupe node ID service with ID %v", badServiceId1))
	}
	if err := checker1.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred waiting for first dupe node ID service with ID %v to start", badServiceId1))
	}
	allServiceIds[badServiceId1] = true

	badServiceClient1, err := castedNetwork.GetGeckoClient(badServiceId1)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred getting the Gecko client for the first dupe node ID service with ID %v", badServiceId1))
	}
	allGeckoClients[badServiceId1] = badServiceClient1

	badServiceNodeId1, err := badServiceClient1.InfoApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get node ID from first dupe node ID service with ID %v", badServiceId1))
	}
	allNodeIds[badServiceId1] = badServiceNodeId1

	logrus.Info("Successfully added first node with soon-to-be-duplicated ID")

	// Verify that the new node got accepted by everyone
	logrus.Infof("Verifying that the new node with service ID %v was accepted by all bootstrappers...", badServiceId1)
	if err := verifyNetworkFullyConnected(allServiceIds, bootServiceIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}
	logrus.Infof("New node with service ID %v was accepted by all bootstrappers", badServiceId1)

	// Now, add a second node with the same ID
	badServiceId2 := 2
	logrus.Infof("Adding second node with service ID %v which will be a duplicated node ID...", badServiceId2)
	checker2, err := castedNetwork.AddService(SAME_CERT_CONFIG_ID, badServiceId2)
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

	badServiceNodeId2, err := badServiceClient2.InfoApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get node ID from first dupe node ID service with ID %v", badServiceId2))
	}
	allNodeIds[badServiceId2] = badServiceNodeId2
	logrus.Info("Second node added, causing duplicate node ID")

	// At this point, it's undefined what happens with the two nodes with duplicate IDs; verify that the original nodes
	//  in the network operate normally amongst themselves
	logrus.Info("Connection behaviour to nodes with duplicate IDs is undefined, so verifying that the original nodes connect as expected...")
	for serviceId, _ := range originalServiceIds {
		acceptableNodeIds := make(map[string]bool)

		// All original nodes should have the boot nodes (though a boot node won't have itself)
		for bootServiceId, _ := range bootServiceIds {
			if serviceId != bootServiceId {
				bootNodeId := allNodeIds[bootServiceId]
				acceptableNodeIds[bootNodeId] = true
			}
		}

		if _, found := bootServiceIds[serviceId]; found {
			// Boot nodes should have the original node, one of the duplicates, and MAY have the duplicate nodes
			acceptableNodeIds[allNodeIds[NODE_SERVICE_ID]] = true
			acceptableNodeIds[badServiceNodeId1] = true
			acceptableNodeIds[badServiceNodeId2] = true
			if err := verifyExpectedPeers(serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(originalServiceIds)-1, true); err != nil {
				context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
			}
		} else {
			// The original non-boot node should have exactly the boot nodes
			if err := verifyExpectedPeers(serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(bootServiceIds), false); err != nil {
				context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
			}
		}
	}
	logrus.Info("Verified that original nodes are still connected to each other")

	// Now, kill the first dupe node to leave only the second (who everyone should connect with)
	logrus.Info("Removing first node with duplicate ID...")
	if err := castedNetwork.RemoveService(badServiceId1); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not remove the first service with duped node ID"))
	}
	delete(allServiceIds, badServiceId1)
	delete(allGeckoClients, badServiceId1)
	delete(allNodeIds, badServiceId1)
	logrus.Info("Successfully removed first node with duplicate ID, leaving only the second")

	// Now that the first duped node is gone, verify that the original node is still connected to just boot nodes and
	//  the second duped-ID node is now accepted by the boot nodes
	logrus.Info("Verifying that the network has connected to the second node with a previously-duplicated node ID...")
	if err := verifyNetworkFullyConnected(allServiceIds, bootServiceIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}
	logrus.Info("Verified that the network has settled on the second node with previously-duplicated ID")
}

func (f StakingNetworkDuplicateNodeIdTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getStakingNetworkLoader(map[int]int{
		NODE_SERVICE_ID:           NORMAL_NODE_CONFIG_ID,
	}, f.imageName)
}

func (f StakingNetworkDuplicateNodeIdTest) GetTimeout() time.Duration {
	return 180 * time.Second
}

// =============== Helper functions =============================

/*
Args:
	desiredServices: Mapping of service_id -> configuration_id for all services *in addition to the boot nodes* that the user wants
 */
func getStakingNetworkLoader(desiredServices map[int]int, imageName string) (testsuite.TestNetworkLoader, error) {
	serviceConfigs := map[int]ava_networks.TestGeckoNetworkServiceConfig{
		NORMAL_NODE_CONFIG_ID: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, imageName, 2, 2),
		SAME_CERT_CONFIG_ID:   *ava_networks.NewTestGeckoNetworkServiceConfig(false, ava_services.LOG_LEVEL_DEBUG, imageName, 2, 2),
	}
	return ava_networks.NewTestGeckoNetworkLoader(
		ava_services.LOG_LEVEL_DEBUG,
		true,
		serviceConfigs,
		desiredServices,
		2,
		2)
}

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

/*
Asserts that the network is fully connected, meaning:
1) The stakers have all the other nodes in the network besides themselves in their peer list
2) All non-stakers have all the stakers in their peer list

Args:
	allServiceIds: All the service IDs in the network, and the IDs that will be iterated over to check
	stakerServiceIds: The service IDs of nodes that we expect to be fully connected - i.e. any node that's actually
		staking. Most of the time this will be just the bootstrappers, but if we add more stakers then this set will
		expand beyond the bootstrappers.
	allNodeIds: The mapping of servcie_id -> node_id
 */
func verifyNetworkFullyConnected(allServiceIds map[int]bool, stakerServiceIds map[int]bool, allNodeIds map[int]string, allGeckoClients map[int]*gecko_client.GeckoClient) error {
	logrus.Tracef("All node IDs in network being verified: %v", allNodeIds)
	for serviceId, _ := range allServiceIds {
		_, isStaker := stakerServiceIds[serviceId]

		acceptableNodeIds := make(map[string]bool)
		for comparisonId, _ := range allServiceIds {
			// Nodes will never have themselves in their peer list
			if serviceId == comparisonId {
				continue
			}
			_, isComparisonStaker := stakerServiceIds[comparisonId]

			// Staker nodes will have all other nodes in their peer list
			// Non-stakers will only have the stakers
			if isStaker || (!isStaker && isComparisonStaker) {
				comparisonNodeId := allNodeIds[comparisonId]
				acceptableNodeIds[comparisonNodeId] = true
			}
		}

		logrus.Debugf("Expecting serviceId %v to have the following peer node IDs, %v", serviceId, acceptableNodeIds)
		if err := verifyExpectedPeers(serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(acceptableNodeIds), false); err != nil {
			return stacktrace.Propagate(err, "An error occurred verifying the expected peers list")
		}
	}
	return nil
}

/*
Verifies that a node's actual peers are what we expect

Args:
	serviceId: Service ID of the node whose peers are being examined
	client: Gecko client for the node being examined
	acceptableNodeIds: A "set" of acceptable node IDs where, if a peer doesn't have this ID, the test will be failed
	expectedNumPeers: The number of peers we expect this node to have
	atLeast: If true, indicates that the number of peers must be AT LEAST the expected number of peers; if false, must be exact
 */
func verifyExpectedPeers(serviceId int, client *gecko_client.GeckoClient, acceptableNodeIds map[string]bool, expectedNumPeers int, atLeast bool) error {
	peers, err := client.InfoApi().GetPeers()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get peers from service with ID %v", serviceId)
	}

	actualNumPeers := len(peers)
	var condition bool
	var operatorAsserted string
	if atLeast {
		condition = actualNumPeers >= expectedNumPeers
		operatorAsserted = ">="
	} else {
		condition = actualNumPeers == expectedNumPeers
		operatorAsserted = "=="
	}

	if !condition {
		return stacktrace.NewError(
			"Service ID %v actual num peers, %v, is not %v expected num peers, %v",
			serviceId,
			actualNumPeers,
			operatorAsserted,
			expectedNumPeers,
		)
	}

	// Verify that IDs of the peers we have are in our list of acceptable IDs
	for _, peer := range peers {
		_, found := acceptableNodeIds[peer.Id]
		if !found {
			return stacktrace.NewError("Service ID %v has a peer with node ID %v that we don't recognize", serviceId, peer.Id)
		}
	}
	return nil
}
