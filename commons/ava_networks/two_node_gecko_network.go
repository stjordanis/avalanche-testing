package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

// ============= Network =====================
type TwoNodeGeckoNetwork struct{
	bootNode      ava_services.GeckoService
	dependentNode ava_services.GeckoService
}
func (network TwoNodeGeckoNetwork) GetBootNode() ava_services.GeckoService {
	return network.bootNode
}
func (network TwoNodeGeckoNetwork) GetDependentNode() ava_services.GeckoService {
	return network.dependentNode
}

// ============= Loader =====================
type TwoNodeGeckoNetworkLoader struct {}
func (loader TwoNodeGeckoNetworkLoader) GetNetworkConfig() (*networks.ServiceNetworkConfig, error) {
	initializerCore := ava_services.NewGeckoServiceInitializerCore(
		2,
		2,
		false,
		ava_services.LOG_LEVEL_DEBUG)
	availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}

	builder := networks.NewServiceNetworkConfigBuilder()
	config1 := builder.AddTestImageConfiguration(initializerCore, availabilityCheckerCore)
	bootNode, err := builder.AddService(config1, make(map[int]bool))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add bootnode service")
	}
	_, err = builder.AddService(
		config1,
		map[int]bool{
			bootNode: true,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add dependent service")
	}
	return builder.Build(), nil
}
func (loader TwoNodeGeckoNetworkLoader) WrapNetwork(services map[int]services.Service) (interface{}, error) {
	return TwoNodeGeckoNetwork{
		bootNode:      services[0].(ava_services.GeckoService),
		dependentNode: services[1].(ava_services.GeckoService),
	}, nil
}

