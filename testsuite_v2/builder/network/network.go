package network

import (
	"fmt"
	"time"

	avalancheNetwork "github.com/ava-labs/avalanche-testing/avalanche/networks"
	avalancheService "github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
)

// Network defines the Network structure of the Nodes in the Topology
type Network struct {
	svcNetwork     *networks.ServiceNetwork
	NodesNames     map[networks.ServiceID]int
	Nodes          map[networks.ConfigurationID]*Node
	snowQuorumSize int
	snowSampleSize int
	isStaking      bool
	image          string
	txFee          uint64
}

// New creates the Network builder
func New() *Network {
	return &Network{
		NodesNames: map[networks.ServiceID]int{},
		Nodes:      map[networks.ConfigurationID]*Node{},
	}
}

func (n *Network) TxFee(txFee uint64) *Network {
	n.txFee = txFee
	return n
}

func (n *Network) SnowSize(snowSampleSize int, snowQuorumSize int) *Network {
	n.snowQuorumSize = snowQuorumSize
	n.snowSampleSize = snowSampleSize
	return n
}

func (n *Network) Image(image string) *Network {
	n.image = image
	return n
}

func (n *Network) IsStaking(isStaking bool) *Network {
	n.isStaking = isStaking
	return n
}

func (n *Network) Generate() (networks.NetworkLoader, error) {

	serviceConfigs := map[networks.ConfigurationID]avalancheNetwork.TestAvalancheNetworkServiceConfig{}
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{}

	for _, node := range n.Nodes {
		serviceConfigs[node.nodeConfigID] = *avalancheNetwork.NewTestAvalancheNetworkServiceConfig(
			node.varyCerts,
			node.serviceLogLevel,
			node.imageName,
			node.snowQuorumSize,
			node.snowSampleSize,
			node.networkInitialTimeout,
			node.additionalCLIArgs,
		)

		desiredServices[node.serviceID] = node.nodeConfigID
	}

	// Return an Avalanche Test Network with this service:configuration mapping.
	return avalancheNetwork.NewTestAvalancheNetworkLoader(
		n.isStaking,
		n.image,
		// TODO change this ?
		avalancheService.DEBUG,
		n.snowQuorumSize,
		n.snowSampleSize,
		n.txFee,
		2*time.Second,
		serviceConfigs,
		desiredServices,
	)
}

func (n *Network) AddNode(node *Node) *Network {

	var nodeNumber int
	if val, ok := n.NodesNames[node.serviceID]; !ok {
		nodeNumber = 0
	} else {
		nodeNumber = val + 1
	}
	n.NodesNames[node.serviceID] = nodeNumber

	n.Nodes[networks.ConfigurationID(fmt.Sprintf("%s-%d", node.serviceID, nodeNumber))] = node
	return n
}

type Node struct {
	serviceID             networks.ServiceID
	nodeConfigID          networks.ConfigurationID
	varyCerts             bool
	serviceLogLevel       avalancheService.AvalancheLogLevel
	imageName             string
	snowQuorumSize        int
	snowSampleSize        int
	networkInitialTimeout time.Duration
	additionalCLIArgs     map[string]string
}

func NewNode(nodeServiceID networks.ServiceID) *Node {
	return &Node{
		serviceID:             nodeServiceID,
		nodeConfigID:          "normal-config",
		varyCerts:             true,
		serviceLogLevel:       avalancheService.DEBUG,
		imageName:             "avaplatform/avalanchego:latest",
		snowQuorumSize:        1,
		snowSampleSize:        1,
		networkInitialTimeout: 2 * time.Second,
		additionalCLIArgs:     make(map[string]string),
	}
}

func (node *Node) Image(imageName string) *Node {
	node.imageName = imageName
	return node
}

func (node *Node) SnowConf(snowQuorumSize int, snowSampleSize int) *Node {
	node.snowQuorumSize = snowQuorumSize
	node.snowSampleSize = snowSampleSize
	return node
}
