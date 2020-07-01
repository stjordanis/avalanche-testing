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
)

// ================ RPC Workflow Test ===================================
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
	stakerNodeId, err := stakerClient.InfoApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get staker node ID."))
	}
	delegatorNodeId, err := stakerClient.InfoApi().GetNodeId()
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
	})
}
func (test StakingNetworkRpcWorkflowTest) GetTimeout() time.Duration {
	return 90 * time.Second
}


// =================== Fully Connected Test ==============================
type FiveNodeStakingNetworkFullyConnectedTest struct{}
func (test FiveNodeStakingNetworkFullyConnectedTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	allServiceIds := castedNetwork.GetAllBootServiceIds()
	allServiceIds[NODE_SERVICE_ID] = true

	extraStakerClient, err := castedNetwork.GetGeckoClient(NODE_SERVICE_ID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get extra staker client."))
	}
	highLevelExtraStakerClient := ava_networks.NewHighLevelGeckoClient(
		extraStakerClient,
		STAKER_USERNAME,
		STAKER_PASSWORD)
	err = highLevelExtraStakerClient.GetFundsAndStartValidating(SEED_AMOUNT, STAKE_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add extra staker."))
	}

	// collect set of IDs in network
	nodeIdSet := map[string]bool{}
	for serviceId, _ := range allServiceIds {
		client, err := castedNetwork.GetGeckoClient(serviceId)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client for service with ID %v", serviceId))
		}
		id, err := client.InfoApi().GetNodeId()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get node ID of service with ID %v", serviceId))
		}
		nodeIdSet[id] = true
	}

	logrus.Debugf("Network ID Set: %+v", nodeIdSet)

	// verify bootstrapper peer lists have full set of IDs in network.
	for serviceId, _ := range allServiceIds {
		client, err := castedNetwork.GetGeckoClient(serviceId)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get client for service with ID %v", serviceId))
		}
		peers, err := client.InfoApi().GetPeers()
		nodeId, err := client.InfoApi().GetNodeId()
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get peers of service with ID %v", serviceId))
		}
		logrus.Debugf("Length of Peer set for node %v: %v", serviceId, len(peers))

		peerSet := map[string]bool{}
		for _, peer := range peers {
			peerSet[peer.Id] = true
		}
		for expectedNodeId, _ := range nodeIdSet {
			// Nodes are expected to have all other nodes in the network on their peer list.
			if nodeId != expectedNodeId {
				context.AssertTrue(peerSet[expectedNodeId], stacktrace.NewError("Didn't find node ID %v in peer set", expectedNodeId))
			}
		}
	}
}

func (test FiveNodeStakingNetworkFullyConnectedTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getStakingNetworkLoader(map[int]int{
		NODE_SERVICE_ID:           NORMAL_NODE_CONFIG_ID,
	})
}

func (test FiveNodeStakingNetworkFullyConnectedTest) GetTimeout() time.Duration {
	return 60 * time.Second
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
	context.AssertTrue(len(validators) >= 1, stacktrace.NewError("No validators found"))
}

func (test FiveNodeStakingNetworkGetValidatorsTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getStakingNetworkLoader(map[int]int{
		NODE_SERVICE_ID:           NORMAL_NODE_CONFIG_ID,
	})
}

func (test FiveNodeStakingNetworkGetValidatorsTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

// =============== Duplicate Node ID Test ==============================
type FiveNodeStakingNetworkDuplicateIdTest struct {}
func (f FiveNodeStakingNetworkDuplicateIdTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	bootServiceIds := castedNetwork.GetAllBootServiceIds()

	allServiceIds := make(map[int]bool)
	allServiceIds[NODE_SERVICE_ID] = true
	for bootServiceId, _ := range bootServiceIds {
		allServiceIds[bootServiceId] = true
	}

	allGeckoClients := make(map[int]*gecko_client.GeckoClient)
	allNodeIds := make(map[int]string)
	for serviceId, _ := range allServiceIds {
		client, err := castedNetwork.GetGeckoClient(serviceId)
		if err != nil {
			context.Fatal(stacktrace.NewError("An error occurred getting the Gecko client for service with ID %v", serviceId))
		}
		allGeckoClients[serviceId] = client
		nodeId, err := client.InfoApi().GetNodeId()
		if err != nil {
			context.Fatal(stacktrace.NewError("An error occurred getting the Gecko node ID for service with ID %v", serviceId))
		}
		allNodeIds[serviceId] = nodeId
	}

	logrus.Info("Verifying that initial network state is as expected...")
	for serviceId, _ := range allServiceIds {
		acceptableNodeIds := make(map[string]bool)
		for iterServiceId, nodeId := range allNodeIds {
			if serviceId != iterServiceId {
				acceptableNodeIds[nodeId] = true
			}
		}
		verifyExpectedPeers(context, serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(allServiceIds) - 1, false)
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
	for serviceId, _ := range allServiceIds {
		acceptableNodeIds := make(map[string]bool)

		// All original nodes should have the boot nodes (though a boot node won't have itself)
		for bootServiceId, _ := range bootServiceIds {
			if serviceId != bootServiceId {
				bootNodeId := allNodeIds[bootServiceId]
				acceptableNodeIds[bootNodeId] = true
			}
		}

		// Boot nodes will also have the other two nodes
		if _, found := bootServiceIds[serviceId]; found {
			acceptableNodeIds[allNodeIds[NODE_SERVICE_ID]] = true
			acceptableNodeIds[badServiceNodeId1] = true
			verifyExpectedPeers(context, serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(allServiceIds) - 1, false)
		} else {
			verifyExpectedPeers(context, serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(bootServiceIds), false)
		}
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
			verifyExpectedPeers(context, serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(originalServiceIds) - 1, true)
		} else {
			// The original non-boot node should have exactly the boot nodes
			verifyExpectedPeers(context, serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(bootServiceIds), false)
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
	for serviceId, _ := range allServiceIds {
		acceptableNodeIds := make(map[string]bool)

		// All nodes should have the boot nodes (though a boot node won't have itself)
		for bootServiceId, _ := range bootServiceIds {
			if serviceId != bootServiceId {
				bootNodeId := allNodeIds[bootServiceId]
				acceptableNodeIds[bootNodeId] = true
			}
		}

		// Boot nodes should have all nodes
		if _, found := bootServiceIds[serviceId]; found {
			acceptableNodeIds[allNodeIds[NODE_SERVICE_ID]] = true
			acceptableNodeIds[badServiceNodeId2] = true
			verifyExpectedPeers(context, serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(allServiceIds) - 1, false)
		} else {
			verifyExpectedPeers(context, serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(bootServiceIds), false)
		}
	}
	logrus.Info("Verified that the network has settled on the second node with previously-duplicated ID")
}

func (f FiveNodeStakingNetworkDuplicateIdTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getStakingNetworkLoader(map[int]int{
		NODE_SERVICE_ID:           NORMAL_NODE_CONFIG_ID,
	})
}

func (f FiveNodeStakingNetworkDuplicateIdTest) GetTimeout() time.Duration {
	return 120 * time.Second
}

// =============== Helper functions =============================

/*
Args:
	desiredServices: Mapping of service_id -> configuration_id for all services *in addition to the boot nodes* that the user wants
 */
func getStakingNetworkLoader(desiredServices map[int]int) (testsuite.TestNetworkLoader, error) {
	serviceConfigs := map[int]ava_networks.TestGeckoNetworkServiceConfig{
		NORMAL_NODE_CONFIG_ID: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG),
		SAME_CERT_CONFIG_ID:   *ava_networks.NewTestGeckoNetworkServiceConfig(false, ava_services.LOG_LEVEL_DEBUG),
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
Verifies that a node's actual peers are what we expect

Args:
	context: Test context (used for failing if there's a problem)
	serviceId: Service ID of the node whose peers are being examined
	client: Gecko client for the node being examined
	acceptableNodeIds: A "set" of acceptable node IDs where, if a peer doesn't have this ID, the test will be failed
	expectedNumPeers: The number of peers we expect this node to have
	atLeast: If true, indicates that the number of peers must be AT LEAST the expected number of peers; if false, must be exact
 */
func verifyExpectedPeers(
			context testsuite.TestContext,
			serviceId int,
			client *gecko_client.GeckoClient,
			acceptableNodeIds map[string]bool,
			expectedNumPeers int,
			atLeast bool) {
	peers, err := client.InfoApi().GetPeers()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get peers from service with ID %v", serviceId))
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
	context.AssertTrue(
		condition,
		stacktrace.NewError(
			"Service ID %v actual num peers, %v, is not %v expected num peers, %v",
			serviceId,
			actualNumPeers,
			operatorAsserted,
			expectedNumPeers,
		),
	)

	// Verify that IDs of the peers we have are in our list of acceptable IDs
	for _, peer := range peers {
		_, found := acceptableNodeIds[peer.Id]
		context.AssertTrue(found, stacktrace.NewError("Service ID %v has a peer with node ID %v that we don't recognize", serviceId, peer.Id))
	}
}
