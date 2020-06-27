package ava_networks

import (
	"bytes"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services/cert_providers"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

const (
	// The config ID that the first boot node will have, with successive boot nodes being incrementally higher
	bootNodeConfigIdStart int = 987654

	// The service ID that the first boot node will have, with successive boot nodes being incrementally higher
	bootNodeServiceIdStart int = 987654
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

func (network TestGeckoNetwork) GetAllBootServiceIds() []int {
	genesisStakers := DefaultLocalNetGenesisConfig.Stakers
	result := make([]int, 0, len(genesisStakers))
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		result = append(result, bootNodeServiceIdStart + i)
	}
	return result
}

// ============= Loader Service Config ====================================
type TestGeckoNetworkServiceConfig struct {
	// Whether the certs used by services with this configuration will be different or not
	varyCerts bool
	serviceLogLevel ava_services.GeckoLogLevel
}

func NewTestGeckoNetworkServiceConfig(
			varyCerts bool,
			serviceLogLevel ava_services.GeckoLogLevel) *TestGeckoNetworkServiceConfig {
	return &TestGeckoNetworkServiceConfig{
		varyCerts: varyCerts,
		serviceLogLevel: serviceLogLevel,
	}
}

// ============== Loader ======================

type TestGeckoNetworkLoader struct{
	bootNodeLogLevel ava_services.GeckoLogLevel
	isStaking       bool
	serviceConfigs  map[int]TestGeckoNetworkServiceConfig
	desiredServiceConfig  	map[int]int
	snowQuorumSize  int
	snowSampleSize  int
}

/*
Creates a new loader to create a TestGeckoNetwork with the specified parameters, transparently handling the creation
of bootstrapper nodes.

NOTE: Bootstrapper nodes will be created automatically, and will show up in the AvailabilityChecker map that gets returned
upon initialization.

Args:
	numNonBootNodes: The number of nodes that the network will start with on top of the boot nodes
	isStaking: Whether the network will have staking enabled
	serviceConfigs: A mapping of service config ID -> information used to launch the service
	desiredServiceConfigs: A map of service_id -> config_id, one per node that this network should start with
	snowQuorumSize: The Snow consensus sample size used for nodes in the network
	snowSampleSize: The Snow consensus quorum size used for nodes in the network
 */
func NewTestGeckoNetworkLoader(
			bootNodeLogLevel ava_services.GeckoLogLevel,
			isStaking bool,
			serviceConfigs map[int]TestGeckoNetworkServiceConfig,
			desiredServiceConfigs map[int]int,
			snowQuorumSize int,
			snowSampleSize int,
			) (*TestGeckoNetworkLoader, error) {
	if len(desiredServiceConfigs) == 0 {
		return nil, stacktrace.NewError("Must specify at least one node!")
	}

	// Defensive copy
	serviceConfigsCopy := make(map[int]TestGeckoNetworkServiceConfig)
	for configId, configParams := range serviceConfigs {
		if configId >= bootNodeConfigIdStart && configId < (bootNodeConfigIdStart + len(DefaultLocalNetGenesisConfig.Stakers)) {
			return nil, stacktrace.NewError("Config ID %v cannot be used as it's being used as a boot node config ID", configId)
		}
		serviceConfigsCopy[configId] = configParams
	}

	// Defensive copy
	desiredServiceConfigsCopy := make(map[int]int)
	for serviceId, configId := range desiredServiceConfigs {
		if serviceId >= bootNodeServiceIdStart && serviceId < (bootNodeServiceIdStart + len(DefaultLocalNetGenesisConfig.Stakers)) {
			return nil, stacktrace.NewError("Service ID %v cannot be used as it's being used as a boot node config ID", serviceId)
		}
		desiredServiceConfigsCopy[serviceId] = configId
	}

	return &TestGeckoNetworkLoader{
		bootNodeLogLevel: bootNodeLogLevel,
		isStaking:       isStaking,
		serviceConfigs:  serviceConfigsCopy,
		desiredServiceConfig: desiredServiceConfigsCopy,
		snowQuorumSize:  snowQuorumSize,
		snowSampleSize:  snowSampleSize,
	}, nil
}

func (loader TestGeckoNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
	localNetGenesisStakers := DefaultLocalNetGenesisConfig.Stakers

	bootNodeIds := make([]string, 0, len(localNetGenesisStakers))
	for _, staker := range DefaultLocalNetGenesisConfig.Stakers {
		bootNodeIds = append(bootNodeIds, staker.NodeID)
	}

	// Add boot node configs
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		configId := bootNodeConfigIdStart + i

		certString := localNetGenesisStakers[i].TlsCert
		keyString := localNetGenesisStakers[i].PrivateKey

		certBytes := bytes.NewBufferString(certString)
		keyBytes := bytes.NewBufferString(keyString)

		initializerCore := ava_services.NewGeckoServiceInitializerCore(
			loader.snowSampleSize,
			loader.snowQuorumSize,
			loader.isStaking,
			bootNodeIds[0:i], // Only the node IDs of the already-started nodes
			cert_providers.NewStaticGeckoCertProvider(*keyBytes, *certBytes),
			loader.bootNodeLogLevel)
		availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}

		if err := builder.AddTestImageConfiguration(configId, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding bootstrapper node with config ID %v", configId)
		}
	}

	// Add user-custom configs
	for configId, configParams := range loader.serviceConfigs {
		certProvider := cert_providers.NewRandomGeckoCertProvider(configParams.varyCerts)
		initializerCore := ava_services.NewGeckoServiceInitializerCore(
			loader.snowSampleSize,
			loader.snowQuorumSize,
			loader.isStaking,
			bootNodeIds,
			certProvider,
			configParams.serviceLogLevel)
		availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}
		if err := builder.AddTestImageConfiguration(configId, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding Gecko node configuration with ID %v", configId)
		}
	}
	return nil
}

/*
Initializes the Gecko test network, spinning up the correct number of bootstrapper nodes and then the user-requested nodes.

NOTE: The resulting AvailabilityChecker map will contain more IDs than the user requested as it will contain boot nodes. The IDs
that these boot nodes are an unspecified implementation detail.
 */
func (loader TestGeckoNetworkLoader) InitializeNetwork(network *networks.ServiceNetwork) (map[int]services.ServiceAvailabilityChecker, error) {
	availabilityCheckers := make(map[int]services.ServiceAvailabilityChecker)

	localNetGenesisStakers := DefaultLocalNetGenesisConfig.Stakers
	bootNodeIds := make([]string, 0, len(localNetGenesisStakers))
	for _, staker := range localNetGenesisStakers {
		bootNodeIds = append(bootNodeIds, staker.NodeID)
	}

	// Add the bootstrapper nodes
	bootstrapperServiceIds := make(map[int]bool)
	for i := 0; i < len(localNetGenesisStakers); i++ {
		configId := bootNodeConfigIdStart + i
		serviceId := bootNodeServiceIdStart + i
		checker, err := network.AddService(configId, serviceId, bootstrapperServiceIds)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding boot node with ID %v and config ID %v", serviceId, configId)
		}
		bootstrapperServiceIds[serviceId] = true
		availabilityCheckers[serviceId] = *checker
	}

	// User-requested nodes

	for serviceId, configId := range loader.desiredServiceConfig {
		checker, err := network.AddService(configId, serviceId, bootstrapperServiceIds)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding non-boot node with ID %v and config ID %v", serviceId, configId)
		}
		availabilityCheckers[serviceId] = *checker
	}
	return availabilityCheckers, nil
}

func (loader TestGeckoNetworkLoader) WrapNetwork(network *networks.ServiceNetwork) (interface{}, error) {
	return TestGeckoNetwork{
		svcNetwork: network,
	}, nil
}
