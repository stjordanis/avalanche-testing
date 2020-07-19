package duplicate_node_id_test

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	normalNodeConfigId = 0
	sameCertConfigId = 1

	nodeServiceId = 0
)

type StakingNetworkDuplicateNodeIdTest struct {
	imageName string
}
func (f StakingNetworkDuplicateNodeIdTest) Run(network networks.Network, context testsuite.TestContext) {
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

func (f StakingNetworkDuplicateNodeIdTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	serviceConfigs := map[int]ava_networks.TestGeckoNetworkServiceConfig{
		normalNodeConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, f.imageName, 2, 2),
		sameCertConfigId:   *ava_networks.NewTestGeckoNetworkServiceConfig(false, ava_services.LOG_LEVEL_DEBUG, f.imageName, 2, 2),
	}
	desiredServices := map[int]int{
		nodeServiceId: normalNodeConfigId,
	}

	return ava_networks.NewTestGeckoNetworkLoader(
		ava_services.LOG_LEVEL_DEBUG,
		true,
		serviceConfigs,
		desiredServices,
		2,
		2)
}

func (f StakingNetworkDuplicateNodeIdTest) GetTimeout() time.Duration {
	return 180 * time.Second
}
