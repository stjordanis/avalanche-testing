package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_default_testnet"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)


// ============== Network ======================
type TestGeckoNetwork struct{
	svcNetwork *networks.ServiceNetwork
}
func (network TestGeckoNetwork) GetGeckoClient(clientId int) (*gecko_client.GeckoClient, error){
	node, err := network.svcNetwork.GetService(clientId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving service node with ID %v", clientId)
	}
	geckoService := node.Service.(ava_services.GeckoService)
	jsonRpcSocket := geckoService.GetJsonRpcSocket()
	return gecko_client.NewGeckoClient(jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort()), nil
}

func (network TestGeckoNetwork) GetNumberOfNodes() int {
	return network.svcNetwork.GetSize()
}


// ============== Loader ======================
type TestGeckoNetworkServiceConfig struct {
	// Whether the certs used by services with this configuration will be different or not
	varyCerts bool
}

type TestGeckoNetworkLoader struct{
	numNodes int
	numBootNodes int
	isStaking bool
	serviceConfigs map[int]TestGeckoNetworkServiceConfig
	snowQuorumSize int
	snowSampleSize int
	serviceLogLevel ava_services.GeckoLogLevel
	bootNodeIds []string
}

/*
Creates a new loader to create a TestGeckoNetwork with the specified parameters; this is probably the loader that most
tests will use

NOTE: some of the parameters to this function (e.g. snowQuorumSize) are specified here for convenience, but
could be pushed to the TestGeckoNetworkServiceConfig since they're actually service-specific
 */
func NewTestGeckoNetworkLoader(
			numNodes int,
			numBootNodes int,
			isStaking bool,
			serviceConfigs map[int]TestGeckoNetworkServiceConfig,
			snowQuroumSize int,
			snowSampleSize int,
			serviceLogLevel ava_services.GeckoLogLevel) (*TestGeckoNetworkLoader, error) {

	if numNodes == 0 {
		return nil, stacktrace.NewError("Must specify at least one node!")
	}

	if numBootNodes > numNodes {
		return nil, stacktrace.NewError("Asked for %v boot nodes but svcNetwork only has %v nodes", numBootNodes, numNodes)
	}
	if isStaking && numBootNodes > 5 {
		return nil, stacktrace.NewError("Staking networks require five or fewer bootnodes.")
	}

	bootNodeIds := make([]string, 0, numBootNodes)
	for _, staker := ava_default_testnet.LocalTestNet.Stakers[0:numBootNodes]

	// Defensive copy
	serviceConfigsCopy := make(map[int]TestGeckoNetworkServiceConfig)
	for key, value := range serviceConfigs {
		serviceConfigsCopy[key] = value
	}

	return &TestGeckoNetworkLoader{
		numNodes:     numNodes,
		numBootNodes: numBootNodes,
		isStaking: isStaking,
		serviceConfigs: serviceConfigsCopy,
		snowQuorumSize: snowQuroumSize,
		snowSampleSize: snowSampleSize,
		serviceLogLevel: serviceLogLevel,
		bootNodeIds: bootNodeIds,
	}, nil
}

func (loader TestGeckoNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
	initializerCores := make(map[int]ava_services.GeckoServiceInitializerCore)
	for configId, configParams := range loader.serviceConfigs {
		initializerCores[configId] = ava_services.NewGeckoServiceInitializerCore(
			loader.snowSampleSize,
			loader.snowQuorumSize,
			loader.isStaking,
			[]string{ava_services.STAKER_1_NODE_ID},
			*ava_services.NewGeckoCertProvider(true),
			ava_services.LOG_LEVEL_DEBUG)
	}
	initializerCore :=
	availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}
	err := builder.AddTestImageConfiguration(geckoServiceConfigId, initializerCore, availabilityCheckerCore)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the Gecko node configuration")
	}

	return nil
}

func (loader TestGeckoNetworkLoader) InitializeNetwork(network *networks.ServiceNetwork) (map[int]services.ServiceAvailabilityChecker, error) {
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

func (loader TestGeckoNetworkLoader) WrapNetwork(network *networks.ServiceNetwork) (interface{}, error) {
	return TestGeckoNetwork{
		svcNetwork: network,
	}, nil
}
