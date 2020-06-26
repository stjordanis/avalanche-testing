package fixed_gecko_network

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

// ============== Network ======================
type FixedGeckoNetwork struct{
	svcNetwork *networks.ServiceNetwork
}
func (network FixedGeckoNetwork) GetGeckoClient(clientId int) (*gecko_client.GeckoClient, error){
	node, err := network.svcNetwork.GetService(clientId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving service node with ID %v", clientId)
	}
	geckoService := node.Service.(ava_services.GeckoService)
	jsonRpcSocket := geckoService.GetJsonRpcSocket()
	return gecko_client.NewGeckoClient(jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort()), nil
}

func (network FixedGeckoNetwork) GetNumberOfNodes() int {
	return network.svcNetwork.GetSize()
}


// ============== Loader ======================
const (
	geckoServiceConfigId = iota
)

type FixedGeckoNetworkLoader struct{
	numNodes int
	numBootNodes int
	isStaking bool
}

func NewFixedGeckoNetworkLoader(numNodes int, numBootNodes int, isStaking bool) (*FixedGeckoNetworkLoader, error) {
	if numBootNodes > numNodes {
		return nil, stacktrace.NewError("Asked for %v boot nodes but svcNetwork only has %v nodes", numBootNodes, numNodes)
	}
	/*
	  TODO Implement more than one bootnode for staking.
	*/
	if isStaking && numBootNodes != 1 {
		return nil, stacktrace.NewError("Staking networks currently require exactly one bootnode.")
	}
	return &FixedGeckoNetworkLoader{
		numNodes:     numNodes,
		numBootNodes: numBootNodes,
		isStaking: isStaking,
	}, nil
}

func (loader FixedGeckoNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
	initializerCore := ava_services.NewGeckoServiceInitializerCore(
		2,
		2,
		loader.isStaking,
		[]string{ava_services.STAKER_1_NODE_ID},
		ava_services.LOG_LEVEL_DEBUG)
	availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}
	err := builder.AddTestImageConfiguration(geckoServiceConfigId, initializerCore, availabilityCheckerCore)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the Gecko node configuration")
	}
	return nil
}

func (loader FixedGeckoNetworkLoader) InitializeNetwork(network *networks.ServiceNetwork) (map[int]services.ServiceAvailabilityChecker, error) {
	bootNodeIds := make(map[int]bool)
	availabilityCheckers := make(map[int]services.ServiceAvailabilityChecker)
	for i := 0; i < loader.numBootNodes; i++ {
		checker, err := network.AddService(geckoServiceConfigId, i, bootNodeIds)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding boot node with ID %v", i)
		}
		bootNodeIds[i] = true
		availabilityCheckers[i] = *checker
	}
	for i := loader.numBootNodes; i < loader.numNodes; i++ {
		checker, err := network.AddService(geckoServiceConfigId, i, bootNodeIds)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding dependent node with ID %v", i)
		}
		availabilityCheckers[i] = *checker
	}
	return availabilityCheckers, nil
}

func (loader FixedGeckoNetworkLoader) WrapNetwork(network *networks.ServiceNetwork) (interface{}, error) {
	return FixedGeckoNetwork{
		svcNetwork: network,
	}, nil
}
