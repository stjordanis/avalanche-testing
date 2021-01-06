package topology

import (
	"fmt"

	"github.com/sirupsen/logrus"

	avalancheNetwork "github.com/ava-labs/avalanche-testing/avalanche/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"
	"github.com/palantir/stacktrace"
)

// Topology defines how the nodes behave/capabilities in the network
type Topology struct {
	network avalancheNetwork.TestAvalancheNetwork
	context *testsuite.TestContext
	genesis *Genesis
	nodes   map[string]*Node
}

// New creates a new instance of the Topology
func New(network networks.Network, context *testsuite.TestContext) *Topology {
	if net, ok := network.(avalancheNetwork.TestAvalancheNetwork); ok {
		return &Topology{
			network: net,
			context: context,
			nodes:   map[string]*Node{},
		}
	}
	context.Fatal(stacktrace.Propagate(fmt.Errorf("network is not avalancheNetwork.TestAvalancheNetwork"), ""))
	return nil
}

// AddNode adds a new now with both PChain and XChain address
func (s *Topology) AddNode(id string, username string, password string) *Topology {
	client, err := s.network.GetAvalancheClient(networks.ServiceID(id))
	if err != nil {
		s.context.Fatal(stacktrace.Propagate(err, "Unable to fetch the Avalanche client"))
		return s
	}
	newNode := newNode(id, username, password, client, s.context).CreateAddress()
	nodeID, err := client.InfoAPI().GetNodeID()
	if err != nil {
		s.context.Fatal(stacktrace.Propagate(err, "Unable to fetch the InfoAPI Node ID"))
		return s
	}

	logrus.Infof("Added Node: %s", nodeID)
	s.nodes[id] = newNode
	return s
}

// AddGenesis creates the Genesis property in the Topology
func (s *Topology) AddGenesis(nodeID string, username string, password string) *Topology {

	client, err := s.network.GetAvalancheClient(networks.ServiceID(nodeID))
	if err != nil {
		s.context.Fatal(stacktrace.Propagate(err, "Unable to fetch the genesis Avalanche client"))
		return s
	}

	s.genesis = newGenesis(nodeID, username, password, client, s.context)
	err = s.genesis.ImportGenesisFunds()
	if err != nil {
		s.context.Fatal(stacktrace.Propagate(err, "Could not get delegator node ID."))
	}

	return s
}

// Genesis returns the Topology Genesis
func (s *Topology) Genesis() *Genesis {
	return s.genesis
}

// Node returns a Node given the [nodeID]
func (s *Topology) Node(nodeID string) *Node {
	return s.nodes[nodeID]
}
