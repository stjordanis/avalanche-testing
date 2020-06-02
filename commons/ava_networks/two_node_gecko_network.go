package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

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

type TwoNodeGeckoNetworkLoader struct {}
func (loader TwoNodeGeckoNetworkLoader) GetNetworkConfig(testImageName string) (*networks.ServiceNetworkConfig, error) {
	factoryConfig := ava_services.NewGeckoServiceFactoryConfig(
		testImageName,
		2,
		2,
		false,
		ava_services.LOG_LEVEL_DEBUG)
	factory := services.NewServiceFactory(factoryConfig)

	builder := networks.NewServiceNetworkConfigBuilder()
	config1 := builder.AddServiceConfiguration(*factory)
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
func (loader TwoNodeGeckoNetworkLoader) LoadNetwork(ipAddrs map[int]string) (interface{}, error) {
	bootNode := ava_services.NewGeckoService(ipAddrs[0])
	dependentNode := ava_services.NewGeckoService(ipAddrs[1])
	return TwoNodeGeckoNetwork{
		bootNode:      *bootNode,
		dependentNode: *dependentNode,
	}, nil
}
