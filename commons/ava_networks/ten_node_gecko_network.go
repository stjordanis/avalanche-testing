package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

// ============== Network ======================
type TenNodeGeckoNetwork struct{
	geckoServices map[int]ava_services.GeckoService
}
func (network TenNodeGeckoNetwork) GetGeckoService(i int) (ava_services.GeckoService, error){
	if i < 0 || i >= len(network.geckoServices) {
		return ava_services.GeckoService{}, stacktrace.NewError("Invalid Gecko service ID")
	}
	// TODO if we're just getting ava_services back from the ServiceConfigBuilder, then how can we make assumptions here??
	service := network.geckoServices[i]
	return service, nil
}

// ============== Loader ======================
type TenNodeGeckoNetworkLoader struct{}
func (loader TenNodeGeckoNetworkLoader) GetNetworkConfig() (*networks.ServiceNetworkConfig, error) {
	initializerCore := ava_services.NewGeckoServiceInitializerCore(
		2,
		2,
		false,
		ava_services.LOG_LEVEL_DEBUG)
	availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}

	builder := networks.NewServiceNetworkConfigBuilder()
	config1 := builder.AddTestImageConfiguration(initializerCore, availabilityCheckerCore)
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
	for i:=3; i < 10; i++ {
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

func (loader TenNodeGeckoNetworkLoader) WrapNetwork(services map[int]services.Service) (interface{}, error) {
	geckoServices := make(map[int]ava_services.GeckoService)
	for serviceId, service := range services {
		geckoServices[serviceId] = service.(ava_services.GeckoService)
	}
	return TenNodeGeckoNetwork{
		geckoServices: geckoServices,
	}, nil
}

