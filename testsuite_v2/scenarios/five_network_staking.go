package scenarios

import (
	"github.com/ava-labs/avalanche-testing/testsuite_v2/builder/network"
	"github.com/ava-labs/avalanche-testing/testsuite_v2/builder/topology"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"
)

// Scenario is a package that allows to use pre-built images + topologies
// should not be initialized directly
//
// TODO needs to be expanded to other scenarios, right now it's locked at the FiveNetworkStaking
type Scenario struct {
	image        string
	nodes        []string
	nodePassword string
	txFee        uint64
}

// NewFiveNetworkStaking creates a new scenario with five nodes network
func NewFiveNetworkStaking(avalancheImage string) *Scenario {
	return &Scenario{
		image:        avalancheImage,
		nodes:        []string{"first", "second", "third", "fourth", "fifth"},
		nodePassword: "MyNameIs!Jeff",
		txFee:        1 * units.Avax,
	}
}

// NewNetwork returns the default Network for FiveNetworkStaking
func (s *Scenario) NewNetwork() *network.Network {
	newNetwork := network.New().
		IsStaking(true).
		Image(s.image).
		SnowSize(3, 3).
		TxFee(s.txFee)

	for _, nodeName := range s.nodes {
		newNetwork.AddNode(network.NewNode(networks.ServiceID(nodeName)).
			Image(s.image).
			SnowConf(3, 3)).
			TxFee(s.txFee)
	}
	return newNetwork
}

// NewTopology returns the default Node Topology for FiveNetworkStaking
// all nodes were genesis'd with 10k AVAX
// 5k in PChain ( 3k staking + 2k not locked)
// 5K in the XChain
func (s *Scenario) NewTopology(network networks.Network, context *testsuite.TestContext) *topology.Topology {

	top := topology.New(network, context)
	var addresses []string
	for _, nodeName := range s.nodes {
		top.AddNode(nodeName, nodeName, s.nodePassword)
		addresses = append(addresses, top.Node(nodeName).XAddress)
	}

	top.AddGenesis("first", "genesis", s.nodePassword).
		Genesis().
		FundXChainAddresses(
			addresses,
			10*units.KiloAvax+s.txFee,
		)

	for _, nodeName := range s.nodes {

		top.Node(nodeName).BecomeValidator(
			10*units.KiloAvax, // the total genesis'd amount
			5*units.KiloAvax,  // Exported from X-> P + Imported P -> X (2x txfee)
			3*units.KiloAvax,
			s.txFee,
		)
	}

	return top
}
