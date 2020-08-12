package verifier

import (
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

/*
Struct containing logic for verifying the state of the network
We attach these functions to a struct even though the struct doesn't have state to avoid a utils class (which
inevitably becomes a mess of unconnected logic), and to categorize the functions around a common purpose.
*/
type NetworkStateVerifier struct{}

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
func (verifier NetworkStateVerifier) VerifyNetworkFullyConnected(
	allServiceIds map[networks.ServiceID]bool,
	stakerServiceIds map[networks.ServiceID]bool,
	allNodeIds map[networks.ServiceID]string,
	allGeckoClients map[networks.ServiceID]*apis.Client,
) error {
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
		if err := verifier.VerifyExpectedPeers(serviceId, allGeckoClients[serviceId], acceptableNodeIds, len(acceptableNodeIds), false); err != nil {
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
func (verifier NetworkStateVerifier) VerifyExpectedPeers(
	serviceId networks.ServiceID,
	client *apis.Client,
	acceptableNodeIds map[string]bool,
	expectedNumPeers int,
	atLeast bool) error {
	peers, err := client.InfoAPI().Peers()
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
		_, found := acceptableNodeIds[peer.ID]
		if !found {
			return stacktrace.NewError("Service ID %v has a peer with node ID %v that we don't recognize", serviceId, peer.ID)
		}
	}
	return nil
}
