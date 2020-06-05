package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

// ============== Network ======================
type NNodeGeckoNetwork struct{
	geckoServices map[int]ava_services.GeckoService
}
func (network NNodeGeckoNetwork) GetGeckoService(i int) (ava_services.GeckoService, error){
	if i < 0 || i >= len(network.geckoServices) {
		return ava_services.GeckoService{}, stacktrace.NewError("Invalid Gecko service ID")
	}
	// TODO if we're just getting ava_services back from the ServiceConfigBuilder, then how can we make assumptions here??
	service := network.geckoServices[i]
	return service, nil
}

// ============== Loader ======================
type NNodeGeckoNetworkLoader struct{
	numNodes int
	numBootNodes int
}

func NewNNodeGeckoNetworkLoader(numNodes int, numBootNodes int) (*NNodeGeckoNetworkLoader, error) {
	if numBootNodes > numNodes {
		return nil, stacktrace.NewError("Asked for %v boot nodes but network only has %v nodes", numBootNodes, numNodes)
	}
	return &NNodeGeckoNetworkLoader{
		numNodes:     numNodes,
		numBootNodes: numBootNodes,
	}, nil
}

func (loader NNodeGeckoNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkConfigBuilder) error {
	initializerCore := ava_services.NewGeckoServiceInitializerCore(
		2,
		2,
		false,
		ava_services.LOG_LEVEL_DEBUG)
	availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}
	serviceCfg := builder.AddTestImageConfiguration(initializerCore, availabilityCheckerCore)

	bootNodeIds := make(map[int]bool)
	for i := 0; i < loader.numBootNodes; i++ {
		// TODO ID-choosing needs to be deterministic!!
		serviceId, err := builder.AddService(serviceCfg, bootNodeIds)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred when adding a boot node")
		}
		bootNodeIds[serviceId] = true
	}
	for i := loader.numBootNodes; i < loader.numNodes; i++ {
		// TODO ID-choosing needs to be deterministic!!
		_, err := builder.AddService(serviceCfg, bootNodeIds)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred when adding a dependent node")
		}
	}
	return nil
}

func (loader NNodeGeckoNetworkLoader) WrapNetwork(services map[int]services.Service) (interface{}, error) {
	geckoServices := make(map[int]ava_services.GeckoService)
	for serviceId, service := range services {
		geckoServices[serviceId] = service.(ava_services.GeckoService)
	}
	return NNodeGeckoNetwork{
		geckoServices: geckoServices,
	}, nil
}

