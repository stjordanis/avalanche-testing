package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

// ============= Network =====================
type SingleNodeGeckoNetwork struct{
	node ava_services.GeckoService
}
func (network SingleNodeGeckoNetwork) GetNode() ava_services.GeckoService {
	return network.node
}

// ============== Loader ======================
type SingleNodeGeckoNetworkLoader struct {}
func (loader SingleNodeGeckoNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkConfigBuilder) error {
	initializerCore := ava_services.NewGeckoServiceInitializerCore(
		1,
		1,
		false,
		ava_services.LOG_LEVEL_DEBUG)
	availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}

	config1 := builder.AddTestImageConfiguration(initializerCore, availabilityCheckerCore)
	_, err := builder.AddService(config1, make(map[int]bool))
	if err != nil {
		return stacktrace.Propagate(err, "Could not add service")
	}
	return nil
}

func (loader SingleNodeGeckoNetworkLoader) WrapNetwork(services map[int]services.Service) (interface{}, error) {
	return SingleNodeGeckoNetwork{
		node: services[0].(ava_services.GeckoService),
	}, nil
}

