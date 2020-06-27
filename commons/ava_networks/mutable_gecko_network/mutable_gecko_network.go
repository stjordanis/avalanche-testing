package mutable_gecko_network

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

// ============== Network ======================
type MutableGeckoNetwork struct{
	svcNetwork *networks.ServiceNetwork
}
func (network MutableGeckoNetwork) GetGeckoClient(clientId int) (*gecko_client.GeckoClient, error){
	node, err := network.svcNetwork.GetService(clientId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving service node with ID %v", clientId)
	}
	geckoService := node.Service.(ava_services.GeckoService)
	jsonRpcSocket := geckoService.GetJsonRpcSocket()
	return gecko_client.NewGeckoClient(jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort()), nil
}

func (network MutableGeckoNetwork) GetNumberOfNodes() int {
	return network.svcNetwork.GetSize()
}


// ============== Loader ======================
const (
	varyingCertGeckoConfigId = iota
	fixedCertGeckoConfigId = iota

	snowSampleSize = 2
	snowQuorumSize = 2
)

type MutableGeckoNetworkLoader struct {
	numNodes int
	numBootNodes int
	isStaking bool
}

func NewMutableGeckoNetworkLoader(initNumNodes int, numBootNodes int, isStaking bool) (*MutableGeckoNetworkLoader, error) {
	if numBootNodes > initNumNodes {
		return nil, stacktrace.NewError("Asked for %v boot nodes but svcNetwork only has %v nodes", numBootNodes, initNumNodes)
	}
	/*
	  TODO Implement more than one bootnode for staking.
	 */
	if isStaking && numBootNodes != 1 {
		return nil, stacktrace.NewError("Staking networks currently require exactly one bootnode.")
	}
	return &MutableGeckoNetworkLoader{
		numNodes:     initNumNodes,
		numBootNodes: numBootNodes,
		isStaking:    isStaking,
	}, nil
}

func (loader MutableGeckoNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
	availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}

	varyingCertInitializerCore := ava_services.NewGeckoServiceInitializerCore(
		snowSampleSize,
		snowQuorumSize,
		loader.isStaking,
		[]string{ava_services.STAKER_1_NODE_ID},
		*ava_services.NewGeckoCertProvider(true),
		ava_services.LOG_LEVEL_DEBUG)
	if err := builder.AddTestImageConfiguration(varyingCertGeckoConfigId, varyingCertInitializerCore, availabilityCheckerCore); err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the varying-cert Gecko node configuration")
	}

	//
	fixedCertInitializerCore := ava_services.NewGeckoServiceInitializerCore(
		snowSampleSize,
		snowQuorumSize,
		loader.isStaking,
		[]string{ava_services.STAKER_1_NODE_ID},
		*ava_services.NewGeckoCertProvider(false),
		ava_services.LOG_LEVEL_DEBUG)
	if err := builder.AddTestImageConfiguration(varyingCertGeckoConfigId, fixedCertInitializerCore, availabilityCheckerCore); err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the fixed-cert Gecko node configuration")
	}

	return nil
}

func (loader MutableGeckoNetworkLoader) InitializeNetwork(network *networks.ServiceNetwork) (map[int]services.ServiceAvailabilityChecker, error) {
	bootNodeIds := make(map[int]bool)
	availabilityCheckers := make(map[int]services.ServiceAvailabilityChecker)
	for i := 0; i < loader.numBootNodes; i++ {
		checker, err := network.AddService(varyingCertGeckoConfigId, i, bootNodeIds)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding boot node with ID %v", i)
		}
		bootNodeIds[i] = true
		availabilityCheckers[i] = *checker
	}
	for i := loader.numBootNodes; i < loader.numNodes; i++ {
		checker, err := network.AddService(varyingCertGeckoConfigId, i, bootNodeIds)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding dependent node with ID %v", i)
		}
		availabilityCheckers[i] = *checker
	}
	return availabilityCheckers, nil
}

func (loader MutableGeckoNetworkLoader) WrapNetwork(network *networks.ServiceNetwork) (interface{}, error) {
	return MutableGeckoNetwork{
		svcNetwork: network,
	}, nil
}
