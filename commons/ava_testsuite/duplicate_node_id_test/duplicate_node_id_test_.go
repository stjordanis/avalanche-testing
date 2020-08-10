package duplicate_node_id_test

import (
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_networks"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_services"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/verifier"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigId networks.ConfigurationID = "normal-config"
	sameCertConfigId networks.ConfigurationID = "same-cert-config"

	vanillaNodeServiceId networks.ServiceID = "vanilla-node"
	badServiceId1 networks.ServiceID = "bad-service-1"
	badServiceId2 networks.ServiceID = "bad-service-2"
)

type DuplicateNodeIdTest struct {
	ImageName string
	Verifier  verifier.NetworkStateVerifier
}

func (test DuplicateNodeIdTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	bootServiceIds := castedNetwork.GetAllBootServiceIds()

	allServiceIds := make(map[networks.ServiceID]bool)
	for bootServiceId, _ := range bootServiceIds {
		allServiceIds[bootServiceId] = true
	}
	allServiceIds[vanillaNodeServiceId] = true

	allNodeIds, allGeckoClients := getNodeIdsAndClients(context, castedNetwork, allServiceIds)
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIds, bootServiceIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}

	// We'll need these later
	originalServiceIds := make(map[networks.ServiceID]bool)
	for serviceId, _ := range allServiceIds {
		originalServiceIds[serviceId] = true
	}

	logrus.Debugf("Service IDs before adding any nodes: %v", allServiceIds)
	logrus.Debugf("Gecko node IDs before adding any nodes: %v", allNodeIds)

	// Add the first dupe node ID (should look normal from a network perspective
	logrus.Info("Adding first node with soon-to-be-duplicated node ID...")
	checker1, err := castedNetwork.AddService(sameCertConfigId, badServiceId1)
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
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIds, bootServiceIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}
	logrus.Infof("New node with service ID %v was accepted by all bootstrappers", badServiceId1)

	// Now, add a second node with the same ID
	logrus.Infof("Adding second node with service ID %v which will be a duplicated node ID...", badServiceId2)
	checker2, err := castedNetwork.AddService(sameCertConfigId, badServiceId2)
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
			acceptableNodeIds[allNodeIds[vanillaNodeServiceId]] = true
			acceptableNodeIds[badServiceNodeId1] = true
			acceptableNodeIds[badServiceNodeId2] = true
			if err := test.Verifier.VerifyExpectedPeers(serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(originalServiceIds)-1, true); err != nil {
				context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
			}
		} else {
			// The original non-boot node should have exactly the boot nodes
			if err := test.Verifier.VerifyExpectedPeers(serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(bootServiceIds), false); err != nil {
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
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIds, bootServiceIds, allNodeIds, allGeckoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}
	logrus.Info("Verified that the network has settled on the second node with previously-duplicated ID")
}

func (test DuplicateNodeIdTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	serviceConfigs := map[networks.ConfigurationID]ava_networks.TestGeckoNetworkServiceConfig{
		normalNodeConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(
			true,
			ava_services.LOG_LEVEL_DEBUG,
			test.ImageName,
			2,
			2,
			make(map[string]string),
		),
		sameCertConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(
			false,
			ava_services.LOG_LEVEL_DEBUG,
			test.ImageName,
			2,
			2,
			make(map[string]string),
		),
	}
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{
		vanillaNodeServiceId: normalNodeConfigId,
	}
	return ava_networks.NewTestGeckoNetworkLoader(
		true,
		test.ImageName,
		ava_services.LOG_LEVEL_DEBUG,
		2,
		2,
		serviceConfigs,
		desiredServices)
}

func (test DuplicateNodeIdTest) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

func (test DuplicateNodeIdTest) GetSetupBuffer() time.Duration {
	// TODO drop this when the availabilityChecker doesn't have a sleep (because we spin up a bunch of nodes before execution)
	return 6 * time.Minute
}

// ================ Helper functions ==================================
/*
This helper function will grab node IDs and Gecko clients
*/
func getNodeIdsAndClients(
	testContext testsuite.TestContext,
	network ava_networks.TestGeckoNetwork,
	allServiceIds map[networks.ServiceID]bool,
) (allNodeIds map[networks.ServiceID]string, allGeckoClients map[networks.ServiceID]*gecko_client.GeckoClient) {
	allGeckoClients = make(map[networks.ServiceID]*gecko_client.GeckoClient)
	allNodeIds = make(map[networks.ServiceID]string)
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
