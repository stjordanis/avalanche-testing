package networks

import (
	"bytes"
	"fmt"
	"time"

	"strconv"
	"strings"

	avalancheService "github.com/ava-labs/avalanche-e2e-tests/avalanche/services"
	"github.com/ava-labs/avalanche-e2e-tests/avalanche/services/certs"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/avalanche-e2e-tests/utils/constants"

	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

const (
	// The prefix for boot node configuration IDs, with an integer appended to specify each one
	bootNodeConfigIDPrefix string = "boot-node-config-"

	// The prefix for boot node service IDs, with an integer appended to specify each one
	bootNodeServiceIDPrefix string = "boot-node-"
)

// ========================================================================================================
//                                    Gecko Test Network
// ========================================================================================================
const (
	containerStopTimeout = 30 * time.Second
)

// TestGeckoNetwork wraps Kurtosis' ServiceNetwork that is meant to be the interface tests use for interacting with Avalanche
// networks of Gecko nodes
type TestGeckoNetwork struct {
	networks.Network

	svcNetwork *networks.ServiceNetwork
}

// GetAvalancheClient returns the API Client for the node with the given service ID
func (network TestGeckoNetwork) GetAvalancheClient(serviceID networks.ServiceID) (*apis.Client, error) {
	node, err := network.svcNetwork.GetService(serviceID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving service node with ID %v", serviceID)
	}
	geckoService := node.Service.(avalancheService.GeckoService)
	jsonRPCSocket := geckoService.GetJSONRPCSocket()
	uri := fmt.Sprintf("http://%s:%d", jsonRPCSocket.GetIpAddr(), jsonRPCSocket.GetPort().Int())
	return apis.NewClient(uri, constants.DefaultRequestTimeout), nil
}

// GetAllBootServiceIDs returns the service IDs of all the boot nodes in the network
func (network TestGeckoNetwork) GetAllBootServiceIDs() map[networks.ServiceID]bool {
	result := make(map[networks.ServiceID]bool)
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		bootID := networks.ServiceID(bootNodeServiceIDPrefix + strconv.Itoa(i))
		result[bootID] = true
	}
	return result
}

// AddService adds a service to the test Gecko network, using the given configuration
// Args:
// 		configurationID: The ID of the configuration to use for the service being added
// 		serviceID: The ID to give the service being added
// Returns:
// 		An availability checker that will return true when teh newly-added service is available
func (network TestGeckoNetwork) AddService(configurationID networks.ConfigurationID, serviceID networks.ServiceID) (*services.ServiceAvailabilityChecker, error) {
	availabilityChecker, err := network.svcNetwork.AddService(configurationID, serviceID, network.GetAllBootServiceIDs())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service with service ID %v, configuration ID %v", serviceID, configurationID)
	}
	return availabilityChecker, nil
}

// RemoveService removes the service with the given service ID from the network
// Args:
// 	serviceID: The ID of the service to remove from the network
func (network TestGeckoNetwork) RemoveService(serviceID networks.ServiceID) error {
	if err := network.svcNetwork.RemoveService(serviceID, containerStopTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service with ID %v", serviceID)
	}
	return nil
}

// ========================================================================================================
//                                    Gecko Service Config
// ========================================================================================================

// TestGeckoNetworkServiceConfig is Gecko-specific layer of abstraction atop Kurtosis' service configurations that makes it a
// bit easier for users to define network service configurations specifically for Gecko nodes
type TestGeckoNetworkServiceConfig struct {
	// Whether the certs used by Gecko services created with this configuration will be different or not (which is used
	//  for testing how the network performs using duplicate node IDs)
	varyCerts bool

	// The log level that the Gecko service should use
	serviceLogLevel avalancheService.GeckoLogLevel

	// The image name that Gecko services started from this configuration should use
	// Used primarily for Byzantine tests but can also test heterogenous Gecko versions, for example.
	imageName string

	// The Snow protocol quroum size that Gecko services started from this configuration should have
	snowQuorumSize int

	// The Snow protocol sample size that Gecko services started from this configuration should have
	snowSampleSize int

	// TODO Make these named parameters, so we don't have an arbitrary bag of extra CLI args!
	// A list of extra CLI args that should be passed to the Gecko services started with this configuration
	additionalCLIArgs map[string]string
}

// NewTestGeckoNetworkServiceConfig creates a new Gecko network service config with the given parameters
// Args:
// 		varyCerts: True if the Gecko services created with this configuration will have differing certs (and therefore
// 			differing node IDs), or the same cert (used for a test to see how the Avalanche network behaves with duplicate node
// 			IDs)
// 		serviceLogLevel: The log level that Gecko services started with this configuration will use
// 		imageName: The name of the Docker image that Gecko services started with this configuration will use
// 		snowQuroumSize: The Snow protocol quorum size that Gecko services started with this configuration will use
// 		snowSampleSize: The Snow protocol sample size that Gecko services started with this configuration will use
// 		cliArgs: A key-value mapping of extra CLI args that will be passed to Gecko services started with this configuration
func NewTestGeckoNetworkServiceConfig(
	varyCerts bool,
	serviceLogLevel avalancheService.GeckoLogLevel,
	imageName string,
	snowQuorumSize int,
	snowSampleSize int,
	additionalCLIArgs map[string]string) *TestGeckoNetworkServiceConfig {
	return &TestGeckoNetworkServiceConfig{
		varyCerts:         varyCerts,
		serviceLogLevel:   serviceLogLevel,
		imageName:         imageName,
		snowQuorumSize:    snowQuorumSize,
		snowSampleSize:    snowSampleSize,
		additionalCLIArgs: additionalCLIArgs,
	}
}

// ========================================================================================================
//                                Gecko Test Network Loader
// ========================================================================================================

// TestGeckoNetworkLoader implements Kurtosis' NetworkLoader interface that's used for creating the test network
// of Gecko services
type TestGeckoNetworkLoader struct {
	// The Docker image that should be used for the Gecko boot nodes
	bootNodeImage string

	// The log level that the Gecko boot nodes should use
	bootNodeLogLevel avalancheService.GeckoLogLevel

	// Whether the nodes that get added to the network (boot node and otherwise) will have staking enabled
	isStaking bool

	// A registry of the service configurations available for use in this network
	serviceConfigs map[networks.ConfigurationID]TestGeckoNetworkServiceConfig

	// A mapping of (service ID) -> (service config ID) for the services that the network will initialize with
	desiredServiceConfig map[networks.ServiceID]networks.ConfigurationID

	// The Snow quorum size that the bootstrapper nodes of the network will use
	bootstrapperSnowQuorumSize int

	// The Snow sample size that the bootstrapper nodes of the network will use
	bootstrapperSnowSampleSize int

	// The fixed transaction fee for the network
	txFee uint64
}

// NewTestGeckoNetworkLoader creates a new loader to create a TestGeckoNetwork with the specified parameters, transparently handling the creation
// of bootstrapper nodes.
// NOTE: Bootstrapper nodes will be created automatically, and will show up in the ServiceAvailabilityChecker map that gets returned
// upon initialization.
// Args:
// 	isStaking: Whether the network will have staking enabled
// 	bootNodeImage: The Docker image that should be used to launch the boot nodes
// 	bootNodeLogLevel: The log level that the boot nodes will launch with
// 	bootstrapperSnowQuorumSize: The Snow consensus sample size used for nodes in the network
// 	bootstrapperSnowSampleSize: The Snow consensus quorum size used for nodes in the network
// 	serviceConfigs: A mapping of service config ID -> config info that the network will provide to the test for use
// 	desiredServiceConfigs: A map of service_id -> config_id, one per node, that this network will initialize with
func NewTestGeckoNetworkLoader(
	isStaking bool,
	bootNodeImage string,
	bootNodeLogLevel avalancheService.GeckoLogLevel,
	bootstrapperSnowQuorumSize int,
	bootstrapperSnowSampleSize int,
	txFee uint64,
	serviceConfigs map[networks.ConfigurationID]TestGeckoNetworkServiceConfig,
	desiredServiceConfigs map[networks.ServiceID]networks.ConfigurationID) (*TestGeckoNetworkLoader, error) {
	// Defensive copy
	serviceConfigsCopy := make(map[networks.ConfigurationID]TestGeckoNetworkServiceConfig)
	for configID, configParams := range serviceConfigs {
		if strings.HasPrefix(string(configID), bootNodeConfigIDPrefix) {
			return nil, stacktrace.NewError("Config ID %v cannot be used because prefix %v is reserved for boot node configurations. Choose a configuration id that does not begin with %v.",
				configID,
				bootNodeConfigIDPrefix,
				bootNodeConfigIDPrefix)
		}
		serviceConfigsCopy[configID] = configParams
	}

	// Defensive copy
	desiredServiceConfigsCopy := make(map[networks.ServiceID]networks.ConfigurationID)
	for serviceID, configID := range desiredServiceConfigs {
		if strings.HasPrefix(string(serviceID), bootNodeServiceIDPrefix) {
			return nil, stacktrace.NewError("Service ID %v cannot be used because prefix %v is reserved for boot node services. Choose a service id that does not begin with %v.",
				serviceID,
				bootNodeServiceIDPrefix,
				bootNodeServiceIDPrefix)
		}
		desiredServiceConfigsCopy[serviceID] = configID
	}

	return &TestGeckoNetworkLoader{
		bootNodeImage:              bootNodeImage,
		bootNodeLogLevel:           bootNodeLogLevel,
		isStaking:                  isStaking,
		serviceConfigs:             serviceConfigsCopy,
		desiredServiceConfig:       desiredServiceConfigsCopy,
		bootstrapperSnowQuorumSize: bootstrapperSnowQuorumSize,
		bootstrapperSnowSampleSize: bootstrapperSnowSampleSize,
		txFee:                      txFee,
	}, nil
}

// ConfigureNetwork defines the netwrok's service configurations to be used
func (loader TestGeckoNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
	localNetGenesisStakers := DefaultLocalNetGenesisConfig.Stakers
	bootNodeIDs := make([]string, 0, len(localNetGenesisStakers))
	for _, staker := range DefaultLocalNetGenesisConfig.Stakers {
		bootNodeIDs = append(bootNodeIDs, staker.NodeID)
	}

	// Add boot node configs
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		configID := networks.ConfigurationID(bootNodeConfigIDPrefix + strconv.Itoa(i))

		certString := localNetGenesisStakers[i].TLSCert
		keyString := localNetGenesisStakers[i].PrivateKey

		certBytes := bytes.NewBufferString(certString)
		keyBytes := bytes.NewBufferString(keyString)

		initializerCore := avalancheService.NewGeckoServiceInitializerCore(
			loader.bootstrapperSnowSampleSize,
			loader.bootstrapperSnowQuorumSize,
			loader.txFee,
			loader.isStaking,
			make(map[string]string), // No additional CLI args for the default network
			bootNodeIDs[0:i],        // Only the node IDs of the already-started nodes
			certs.NewStaticGeckoCertProvider(*keyBytes, *certBytes),
			loader.bootNodeLogLevel,
		)
		availabilityCheckerCore := avalancheService.GeckoServiceAvailabilityCheckerCore{}

		if err := builder.AddConfiguration(configID, loader.bootNodeImage, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding bootstrapper node with config ID %v", configID)
		}
	}

	// Add user-custom configs
	for configID, configParams := range loader.serviceConfigs {
		certProvider := certs.NewRandomGeckoCertProvider(configParams.varyCerts)
		imageName := configParams.imageName

		initializerCore := avalancheService.NewGeckoServiceInitializerCore(
			configParams.snowSampleSize,
			configParams.snowQuorumSize,
			loader.txFee,
			loader.isStaking,
			configParams.additionalCLIArgs,
			bootNodeIDs,
			certProvider,
			configParams.serviceLogLevel,
		)
		availabilityCheckerCore := avalancheService.GeckoServiceAvailabilityCheckerCore{}
		if err := builder.AddConfiguration(configID, imageName, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding Gecko node configuration with ID %v", configID)
		}
	}
	return nil
}

// InitializeNetwork implements networks.NetworkLoader that initializes the Gecko test network to the state specified at
// construction time, spinning up the correct number of bootstrapper nodes and subsequently the user-requested nodes.
// NOTE: The resulting services.ServiceAvailabilityChecker map will contain more IDs than the user requested as it will
// 		contain boot nodes. The IDs that these boot nodes are an unspecified implementation detail.
func (loader TestGeckoNetworkLoader) InitializeNetwork(network *networks.ServiceNetwork) (map[networks.ServiceID]services.ServiceAvailabilityChecker, error) {
	availabilityCheckers := make(map[networks.ServiceID]services.ServiceAvailabilityChecker)

	// Add the bootstrapper nodes
	bootstrapperServiceIDs := make(map[networks.ServiceID]bool)
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		configID := networks.ConfigurationID(bootNodeConfigIDPrefix + strconv.Itoa(i))
		serviceID := networks.ServiceID(bootNodeServiceIDPrefix + strconv.Itoa(i))
		checker, err := network.AddService(configID, serviceID, bootstrapperServiceIDs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding boot node with ID %v and config ID %v", serviceID, configID)
		}

		// TODO the first node should have zero dependencies and the rest should
		// have only the first node as a dependency
		bootstrapperServiceIDs[serviceID] = true
		availabilityCheckers[serviceID] = *checker
	}

	// Additional user defined nodes
	for serviceID, configID := range loader.desiredServiceConfig {
		checker, err := network.AddService(configID, serviceID, bootstrapperServiceIDs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding non-boot node with ID %v and config ID %v", serviceID, configID)
		}
		availabilityCheckers[serviceID] = *checker
	}
	return availabilityCheckers, nil
}

// WrapNetwork implements a networks.NetworkLoader function and wraps the underlying networks.ServiceNetwork with the TestGeckoNetwork
func (loader TestGeckoNetworkLoader) WrapNetwork(network *networks.ServiceNetwork) (networks.Network, error) {
	return TestGeckoNetwork{
		svcNetwork: network,
	}, nil
}
