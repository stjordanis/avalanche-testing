package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

// ============== Network ======================
type NNodeGeckoNetwork struct{
	geckoClients map[int]gecko_client.GeckoClient
}
func (network NNodeGeckoNetwork) GetGeckoClient(i int) (gecko_client.GeckoClient, error){
	if i < 0 || i >= len(network.geckoClients) {
		return gecko_client.GeckoClient{}, stacktrace.NewError("Invalid Gecko client ID")
	}
	// TODO if we're just getting ava_services back from the ServiceConfigBuilder, then how can we make assumptions here??
	client := network.geckoClients[i]
	return client, nil
}

// ============== Loader ======================
type NNodeGeckoNetworkLoader struct{
	numNodes int
	numBootNodes int
	isStaking bool
}

func NewNNodeGeckoNetworkLoader(numNodes int, numBootNodes int, isStaking bool) (*NNodeGeckoNetworkLoader, error) {
	if numBootNodes > numNodes {
		return nil, stacktrace.NewError("Asked for %v boot nodes but network only has %v nodes", numBootNodes, numNodes)
	}
	/*
	  TODO Implement more than one bootnode for staking.
	 */
	if isStaking && numBootNodes != 1 {
		return nil, stacktrace.NewError("Staking networks currently require exactly one bootnode.")
	}
	return &NNodeGeckoNetworkLoader{
		numNodes:     numNodes,
		numBootNodes: numBootNodes,
		isStaking: isStaking,
	}, nil
}

func (loader NNodeGeckoNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkConfigBuilder) error {
	initializerCore := ava_services.NewGeckoServiceInitializerCore(
		2,
		2,
		loader.isStaking,
		ava_services.LOG_LEVEL_DEBUG)
	availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}
	serviceCfg := builder.AddTestImageConfiguration(initializerCore, availabilityCheckerCore)

	bootNodeIds := make(map[int]bool)
	for i := 0; i < loader.numBootNodes; i++ {
		// TODO ID-choosing needs to be deterministic!!
		serviceId, err := builder.AddService(serviceCfg, i, bootNodeIds)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred when adding a boot node")
		}
		bootNodeIds[serviceId] = true
	}
	for i := loader.numBootNodes; i < loader.numNodes; i++ {
		// TODO ID-choosing needs to be deterministic!!
		_, err := builder.AddService(serviceCfg, i, bootNodeIds)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred when adding a dependent node")
		}
	}
	return nil
}

func (loader NNodeGeckoNetworkLoader) WrapNetwork(services map[int]services.Service) (interface{}, error) {
	geckoClients := make(map[int]gecko_client.GeckoClient)
	for serviceId, service := range services {
		jsonRpcSocket := service.(ava_services.GeckoService).GetJsonRpcSocket()
		clientPtr := gecko_client.NewGeckoClient(jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort())
		geckoClients[serviceId] = *clientPtr
	}
	return NNodeGeckoNetwork{
		geckoClients: geckoClients,
	}, nil
}

