package ava_networks

import (
	"bytes"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services/cert_providers"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
	"strconv"
	"strings"
	"time"
)

const (
	// The config ID that the first boot node will have, with successive boot nodes being incrementally higher
	bootNodeConfigIdStart int = 987654

	// The service ID that the first boot node will have, with successive boot nodes being incrementally higher
	bootNodeServiceIdPrefix string = "boot-node-"
)

// ============== Network ======================
const (
	containerStopTimeout = 30 * time.Second
)
type TestGeckoNetwork struct{
	networks.Network

	svcNetwork *networks.ServiceNetwork
}
func (network TestGeckoNetwork) GetGeckoClient(serviceId networks.ServiceID) (*gecko_client.GeckoClient, error){
	node, err := network.svcNetwork.GetService(serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving service node with ID %v", serviceId)
	}
	geckoService := node.Service.(ava_services.GeckoService)
	jsonRpcSocket := geckoService.GetJsonRpcSocket()
	return gecko_client.NewGeckoClient(jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort()), nil
}

func (network TestGeckoNetwork) GetAllBootServiceIds() map[networks.ServiceID]bool {
	result := make(map[networks.ServiceID]bool)
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		bootId := networks.ServiceID(bootNodeServiceIdPrefix + strconv.Itoa(i))
		result[bootId] = true
	}
	return result
}

func (network TestGeckoNetwork) AddService(configurationId networks.ConfigurationID, serviceId networks.ServiceID) (*services.ServiceAvailabilityChecker, error) {
	availabilityChecker, err := network.svcNetwork.AddService(configurationId, serviceId, network.GetAllBootServiceIds())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service with service ID %v, configuration ID %v", serviceId, configurationId)
	}
	return availabilityChecker, nil
}

func (network TestGeckoNetwork) RemoveService(serviceId networks.ServiceID) error {
	if err := network.svcNetwork.RemoveService(serviceId, containerStopTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service with ID %v", serviceId)
	}
	return nil
}

// ============= Loader Service Config ====================================
type TestGeckoNetworkServiceConfig struct {
	// Whether the certs used by services with this configuration will be different or not
	varyCerts bool
	serviceLogLevel ava_services.GeckoLogLevel
	// Used primarily for Byzantine tests but can also test heterogenous Gecko versions, for example.
	imageName      string
	snowQuorumSize int
	snowSampleSize int
}

func NewTestGeckoNetworkServiceConfig(
			varyCerts bool,
			serviceLogLevel ava_services.GeckoLogLevel,
			imageName string,
			snowQuorumSize int,
			snowSampleSize int) *TestGeckoNetworkServiceConfig {
	return &TestGeckoNetworkServiceConfig{
		varyCerts:       varyCerts,
		serviceLogLevel: serviceLogLevel,
		imageName:       imageName,
		snowQuorumSize:  snowQuorumSize,
		snowSampleSize:  snowSampleSize,
	}
}

// ============== Loader ======================

type TestGeckoNetworkLoader struct{
	bootNodeImage			   string
	bootNodeLogLevel           ava_services.GeckoLogLevel
	isStaking                  bool
	serviceConfigs             map[networks.ConfigurationID]TestGeckoNetworkServiceConfig
	desiredServiceConfig       map[networks.ServiceID]networks.ConfigurationID
	bootstrapperSnowQuorumSize int
	bootstrapperSnowSampleSize int
}

/*
Creates a new loader to create a TestGeckoNetwork with the specified parameters, transparently handling the creation
of bootstrapper nodes.

NOTE: Bootstrapper nodes will be created automatically, and will show up in the AvailabilityChecker map that gets returned
upon initialization.

Args:
	isStaking: Whether the network will have staking enabled
	bootNodeImage: The Docker image that should be used to launch the boot nodes
	bootNodeLogLevel: The log level that the boot nodes will launch with
	bootstrapperSnowQuorumSize: The Snow consensus sample size used for nodes in the network
	bootstrapperSnowSampleSize: The Snow consensus quorum size used for nodes in the network
	serviceConfigs: A mapping of service config ID -> information used to launch the service
	desiredServiceConfigs: A map of service_id -> config_id, one per node that this network should start with
 */
func NewTestGeckoNetworkLoader(
			isStaking bool,
			bootNodeImage string,
			bootNodeLogLevel ava_services.GeckoLogLevel,
			bootstrapperSnowQuorumSize int,
			bootstrapperSnowSampleSize int,
			serviceConfigs map[networks.ConfigurationID]TestGeckoNetworkServiceConfig,
			desiredServiceConfigs map[networks.ServiceID]networks.ConfigurationID) (*TestGeckoNetworkLoader, error) {
	if len(desiredServiceConfigs) == 0 {
		return nil, stacktrace.NewError("Must specify at least one node!")
	}

	// Defensive copy
	serviceConfigsCopy := make(map[networks.ConfigurationID]TestGeckoNetworkServiceConfig)
	for configId, configParams := range serviceConfigs {
		if int(configId) >= bootNodeConfigIdStart && int(configId) < (bootNodeConfigIdStart + len(DefaultLocalNetGenesisConfig.Stakers)) {
			return nil, stacktrace.NewError("Config ID %v cannot be used as it's being used as a boot node config ID", configId)
		}
		serviceConfigsCopy[configId] = configParams
	}

	// Defensive copy
	desiredServiceConfigsCopy := make(map[networks.ServiceID]networks.ConfigurationID)
	for serviceId, configId := range desiredServiceConfigs {
		if strings.HasPrefix(string(serviceId), bootNodeServiceIdPrefix) {
			return nil, stacktrace.NewError("Service ID %v cannot be used because prefix %v is reserved for boot nodes.", serviceId, bootNodeServiceIdPrefix)
		}
		desiredServiceConfigsCopy[serviceId] = configId
	}

	return &TestGeckoNetworkLoader{
		bootNodeImage: 				bootNodeImage,
		bootNodeLogLevel:           bootNodeLogLevel,
		isStaking:                  isStaking,
		serviceConfigs:             serviceConfigsCopy,
		desiredServiceConfig:       desiredServiceConfigsCopy,
		bootstrapperSnowQuorumSize: bootstrapperSnowQuorumSize,
		bootstrapperSnowSampleSize: bootstrapperSnowSampleSize,
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
		configId := networks.ConfigurationID(bootNodeConfigIdStart + i)

		certString := localNetGenesisStakers[i].TlsCert
		keyString := localNetGenesisStakers[i].PrivateKey

		certBytes := bytes.NewBufferString(certString)
		keyBytes := bytes.NewBufferString(keyString)

		initializerCore := ava_services.NewGeckoServiceInitializerCore(
			loader.bootstrapperSnowSampleSize,
			loader.bootstrapperSnowQuorumSize,
			loader.isStaking,
			bootNodeIds[0:i], // Only the node IDs of the already-started nodes
			cert_providers.NewStaticGeckoCertProvider(*keyBytes, *certBytes),
			loader.bootNodeLogLevel)
		availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}

		if err := builder.AddConfiguration(configId, loader.bootNodeImage, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding bootstrapper node with config ID %v", configId)
		}
	}

	// Add user-custom configs
	for configId, configParams := range loader.serviceConfigs {
		certProvider := cert_providers.NewRandomGeckoCertProvider(configParams.varyCerts)
		imageName := configParams.imageName
		initializerCore := ava_services.NewGeckoServiceInitializerCore(
			configParams.snowSampleSize,
			configParams.snowQuorumSize,
			loader.isStaking,
			bootNodeIds,
			certProvider,
			configParams.serviceLogLevel)
		availabilityCheckerCore := ava_services.GeckoServiceAvailabilityCheckerCore{}
		if err := builder.AddConfiguration(configId, imageName, initializerCore, availabilityCheckerCore); err != nil {
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
func (loader TestGeckoNetworkLoader) InitializeNetwork(network *networks.ServiceNetwork) (map[networks.ServiceID]services.ServiceAvailabilityChecker, error) {
	availabilityCheckers := make(map[networks.ServiceID]services.ServiceAvailabilityChecker)

	// Add the bootstrapper nodes
	bootstrapperServiceIds := make(map[networks.ServiceID]bool)
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		configId := networks.ConfigurationID(bootNodeConfigIdStart + i)
		serviceId := networks.ServiceID(bootNodeServiceIdPrefix + strconv.Itoa(i))
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

func (loader TestGeckoNetworkLoader) WrapNetwork(network *networks.ServiceNetwork) (networks.Network, error) {
	return TestGeckoNetwork{
		svcNetwork: network,
	}, nil
}
