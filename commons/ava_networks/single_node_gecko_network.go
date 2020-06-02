package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

type SingleNodeGeckoNetwork struct{
	node ava_services.GeckoService
}
func (network SingleNodeGeckoNetwork) GetNode() ava_services.GeckoService {
	return network.node
}

type SingleNodeGeckoNetworkLoader struct {}
func (loader SingleNodeGeckoNetworkLoader) GetNetworkConfig(testImageName string) (*networks.ServiceNetworkConfig, error) {
	factoryConfig := ava_services.NewGeckoServiceFactoryConfig(
		testImageName,
		1,
		1,
		false,
		ava_services.LOG_LEVEL_DEBUG)
	factory := services.NewServiceFactory(factoryConfig)

	builder := networks.NewServiceNetworkConfigBuilder()
	config1 := builder.AddServiceConfiguration(*factory)
	_, err := builder.AddService(config1, make(map[int]bool))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add service")
	}
	return builder.Build(), nil
}
func (loader SingleNodeGeckoNetworkLoader) LoadNetwork(ipAddrs map[int]string) (interface{}, error) {
	return SingleNodeGeckoNetwork{
		node: *ava_services.NewGeckoService(ipAddrs[0]),
	}, nil
}
