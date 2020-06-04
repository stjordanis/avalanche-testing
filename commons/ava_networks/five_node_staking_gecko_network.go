package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

type FiveNodeGeckoNetwork struct{
	geckoServices map[int]ava_services.GeckoService
}

func (network FiveNodeGeckoNetwork) GetGeckoService(i int) (ava_services.GeckoService, error){
	if i < 0 || i >= len(network.geckoServices) {
		return ava_services.GeckoService{}, stacktrace.NewError("Invalid Gecko service ID")
	}
	// TODO if we're just getting ava_services back from the ServiceConfigBuilder, then how can we make assumptions here??
	service := network.geckoServices[i]
	return service, nil
}


type FiveNodeGeckoNetworkLoader struct{}
func (loader FiveNodeGeckoNetworkLoader) GetNetworkConfig(testImageName string) (*networks.ServiceNetworkConfig, error) {
	factoryConfig := ava_services.NewGeckoServiceFactoryConfig(
		testImageName,
		2,
		2,
		true,
		ava_services.LOG_LEVEL_DEBUG)
	factory := services.NewServiceFactory(factoryConfig)

	builder := networks.NewServiceNetworkConfigBuilder()
	config1 := builder.AddServiceConfiguration(*factory)
	bootNode0, err := builder.AddService(config1, make(map[int]bool))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add bootnode service")
	}
	bootNode1, err := builder.AddService(
		config1,
		map[int]bool{
			bootNode0: true,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add dependent service")
	}
	bootNode2, err := builder.AddService(
		config1,
		map[int]bool{
			bootNode0: true,
			bootNode1: true,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add dependent service")
	}
	bootNodeMap := map[int]bool{
		bootNode0: true,
		bootNode1: true,
		bootNode2: true,
	}
	for i:=3; i < 5; i++ {
		_, err := builder.AddService(
			config1,
			bootNodeMap,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not add dependent service")
		}
	}

	return builder.Build(), nil
}

func (loader FiveNodeGeckoNetworkLoader) LoadNetwork(ipAddrs map[int]string) (interface{}, error) {
	geckoServices := make(map[int]ava_services.GeckoService)
	for serviceId, ipAddr := range ipAddrs {
		geckoServices[serviceId] = *ava_services.NewGeckoService(ipAddr)
	}
	return TenNodeGeckoNetwork{
		geckoServices: geckoServices,
	}, nil
}
