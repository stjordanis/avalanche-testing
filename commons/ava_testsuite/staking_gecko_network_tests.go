package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
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

/*
Args:
	desiredServices: Mapping of service_id -> configuration_id for all services *in addition to the boot nodes* that the user wants
 */
func getStakingNetworkLoader(desiredServices map[int]int, imageName string) (networks.NetworkLoader, error) {
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
