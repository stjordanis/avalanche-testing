package networks

import (
	"bytes"
	"fmt"
	"time"

	"strconv"
	"strings"

	avalancheService "github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/avalanche/services/certs"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis"
	"github.com/ava-labs/avalanche-testing/utils/constants"

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
//                                    Avalanche Test Network
// ========================================================================================================
const (
	containerStopTimeout = 30 * time.Second
)

// TestAvalancheNetwork wraps Kurtosis' ServiceNetwork that is meant to be the interface tests use for interacting with Avalanche
// networks
type TestAvalancheNetwork struct {
	networks.Network

	svcNetwork *networks.ServiceNetwork
}

// GetAvalancheClient returns the API Client for the node with the given service ID
func (network TestAvalancheNetwork) GetAvalancheClient(serviceID networks.ServiceID) (*apis.Client, error) {
	node, err := network.svcNetwork.GetService(serviceID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving service node with ID %v", serviceID)
	}
	avalancheService := node.Service.(avalancheService.AvalancheService)
	jsonRPCSocket := avalancheService.GetJSONRPCSocket()
	uri := fmt.Sprintf("http://%s:%d", jsonRPCSocket.GetIpAddr(), jsonRPCSocket.GetPort().Int())
	return apis.NewClient(uri, constants.DefaultRequestTimeout), nil
}

// GetAllBootServiceIDs returns the service IDs of all the boot nodes in the network
func (network TestAvalancheNetwork) GetAllBootServiceIDs() map[networks.ServiceID]bool {
	result := make(map[networks.ServiceID]bool)
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		bootID := networks.ServiceID(bootNodeServiceIDPrefix + strconv.Itoa(i))
		result[bootID] = true
	}
	return result
}

// AddService adds a service to the test Avalanche network, using the given configuration
// Args:
// 		configurationID: The ID of the configuration to use for the service being added
// 		serviceID: The ID to give the service being added
// Returns:
// 		An availability checker that will return true when teh newly-added service is available
func (network TestAvalancheNetwork) AddService(configurationID networks.ConfigurationID, serviceID networks.ServiceID) (*services.ServiceAvailabilityChecker, error) {
	availabilityChecker, err := network.svcNetwork.AddService(configurationID, serviceID, network.GetAllBootServiceIDs())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service with service ID %v, configuration ID %v", serviceID, configurationID)
	}
	return availabilityChecker, nil
}

// RemoveService removes the service with the given service ID from the network
// Args:
// 	serviceID: The ID of the service to remove from the network
func (network TestAvalancheNetwork) RemoveService(serviceID networks.ServiceID) error {
	if err := network.svcNetwork.RemoveService(serviceID, containerStopTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service with ID %v", serviceID)
	}
	return nil
}

// ========================================================================================================
//                                    Avalanche Service Config
// ========================================================================================================

// TestAvalancheNetworkServiceConfig is Avalanche-specific layer of abstraction atop Kurtosis' service configurations that makes it a
// bit easier for users to define network service configurations specifically for Avalanche nodes
type TestAvalancheNetworkServiceConfig struct {
	// Whether the certs used by Avalanche services created with this configuration will be different or not (which is used
	//  for testing how the network performs using duplicate node IDs)
	varyCerts bool

	// The log level that the Avalanche service should use
	serviceLogLevel avalancheService.AvalancheLogLevel

	// The image name that Avalanche services started from this configuration should use
	// Used primarily for Byzantine tests but can also test heterogenous Avalanche versions, for example.
	imageName string

	// The Snow protocol quroum size that Avalanche services started from this configuration should have
	snowQuorumSize int

	// The Snow protocol sample size that Avalanche services started from this configuration should have
	snowSampleSize int

	networkInitialTimeout time.Duration

	// TODO Make these named parameters, so we don't have an arbitrary bag of extra CLI args!
	// A list of extra CLI args that should be passed to the Avalanche services started with this configuration
	additionalCLIArgs map[string]string
}

// NewTestAvalancheNetworkServiceConfig creates a new Avalanche network service config with the given parameters
// Args:
// 		varyCerts: True if the Avalanche services created with this configuration will have differing certs (and therefore
// 			differing node IDs), or the same cert (used for a test to see how the Avalanche network behaves with duplicate node
// 			IDs)
// 		serviceLogLevel: The log level that Avalanche services started with this configuration will use
// 		imageName: The name of the Docker image that Avalanche services started with this configuration will use
// 		snowQuroumSize: The Snow protocol quorum size that Avalanche services started with this configuration will use
// 		snowSampleSize: The Snow protocol sample size that Avalanche services started with this configuration will use
// 		cliArgs: A key-value mapping of extra CLI args that will be passed to Avalanche services started with this configuration
func NewTestAvalancheNetworkServiceConfig(
	varyCerts bool,
	serviceLogLevel avalancheService.AvalancheLogLevel,
	imageName string,
	snowQuorumSize int,
	snowSampleSize int,
	networkInitialTimeout time.Duration,
	additionalCLIArgs map[string]string) *TestAvalancheNetworkServiceConfig {
	return &TestAvalancheNetworkServiceConfig{
		varyCerts:             varyCerts,
		serviceLogLevel:       serviceLogLevel,
		imageName:             imageName,
		snowQuorumSize:        snowQuorumSize,
		snowSampleSize:        snowSampleSize,
		networkInitialTimeout: networkInitialTimeout,
		additionalCLIArgs:     additionalCLIArgs,
	}
}

// ========================================================================================================
//                                Avalanche Test Network Loader
// ========================================================================================================

// TestAvalancheNetworkLoader implements Kurtosis' NetworkLoader interface that's used for creating the test network
// of Avalanche services
type TestAvalancheNetworkLoader struct {
	// The Docker image that should be used for the Avalanche boot nodes
	bootNodeImage string

	// The log level that the Avalanche boot nodes should use
	bootNodeLogLevel avalancheService.AvalancheLogLevel

	// Whether the nodes that get added to the network (boot node and otherwise) will have staking enabled
	isStaking bool

	// A registry of the service configurations available for use in this network
	serviceConfigs map[networks.ConfigurationID]TestAvalancheNetworkServiceConfig

	// A mapping of (service ID) -> (service config ID) for the services that the network will initialize with
	desiredServiceConfig map[networks.ServiceID]networks.ConfigurationID

	// The Snow quorum size that the bootstrapper nodes of the network will use
	bootstrapperSnowQuorumSize int

	// The Snow sample size that the bootstrapper nodes of the network will use
	bootstrapperSnowSampleSize int

	// The fixed transaction fee for the network
	txFee uint64

	// The initial timeout for the network
	networkInitialTimeout time.Duration
}

// NewTestAvalancheNetworkLoader creates a new loader to create a TestAvalancheNetwork with the specified parameters, transparently handling the creation
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
func NewTestAvalancheNetworkLoader(
	isStaking bool,
	bootNodeImage string,
	bootNodeLogLevel avalancheService.AvalancheLogLevel,
	bootstrapperSnowQuorumSize int,
	bootstrapperSnowSampleSize int,
	txFee uint64,
	networkInitialTimeout time.Duration,
	serviceConfigs map[networks.ConfigurationID]TestAvalancheNetworkServiceConfig,
	desiredServiceConfigs map[networks.ServiceID]networks.ConfigurationID) (*TestAvalancheNetworkLoader, error) {
	// Defensive copy
	serviceConfigsCopy := make(map[networks.ConfigurationID]TestAvalancheNetworkServiceConfig)
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

	return &TestAvalancheNetworkLoader{
		bootNodeImage:              bootNodeImage,
		bootNodeLogLevel:           bootNodeLogLevel,
		isStaking:                  isStaking,
		serviceConfigs:             serviceConfigsCopy,
		desiredServiceConfig:       desiredServiceConfigsCopy,
		bootstrapperSnowQuorumSize: bootstrapperSnowQuorumSize,
		bootstrapperSnowSampleSize: bootstrapperSnowSampleSize,
		txFee:                      txFee,
		networkInitialTimeout:      networkInitialTimeout,
	}, nil
}

// ConfigureNetwork defines the netwrok's service configurations to be used
func (loader TestAvalancheNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
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

		initializerCore := avalancheService.NewAvalancheServiceInitializerCore(
			loader.bootstrapperSnowSampleSize,
			loader.bootstrapperSnowQuorumSize,
			loader.txFee,
			loader.isStaking,
			loader.networkInitialTimeout,
			make(map[string]string), // No additional CLI args for the default network
			bootNodeIDs[0:i],        // Only the node IDs of the already-started nodes
			certs.NewStaticAvalancheCertProvider(*keyBytes, *certBytes),
			loader.bootNodeLogLevel,
		)
		availabilityCheckerCore := avalancheService.AvalancheServiceAvailabilityCheckerCore{}

		if err := builder.AddConfiguration(configID, loader.bootNodeImage, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding bootstrapper node with config ID %v", configID)
		}
	}

	// Add user-custom configs
	for configID, configParams := range loader.serviceConfigs {
		certProvider := certs.NewRandomAvalancheCertProvider(configParams.varyCerts)
		imageName := configParams.imageName

		initializerCore := avalancheService.NewAvalancheServiceInitializerCore(
			configParams.snowSampleSize,
			configParams.snowQuorumSize,
			loader.txFee,
			loader.isStaking,
			configParams.networkInitialTimeout,
			configParams.additionalCLIArgs,
			bootNodeIDs,
			certProvider,
			configParams.serviceLogLevel,
		)
		availabilityCheckerCore := avalancheService.AvalancheServiceAvailabilityCheckerCore{}
		if err := builder.AddConfiguration(configID, imageName, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding Avalanche node configuration with ID %v", configID)
		}
	}
	return nil
}

// InitializeNetwork implements networks.NetworkLoader that initializes the Avalanche test network to the state specified at
// construction time, spinning up the correct number of bootstrapper nodes and subsequently the user-requested nodes.
// NOTE: The resulting services.ServiceAvailabilityChecker map will contain more IDs than the user requested as it will
// 		contain boot nodes. The IDs that these boot nodes are an unspecified implementation detail.
func (loader TestAvalancheNetworkLoader) InitializeNetwork(network *networks.ServiceNetwork) (map[networks.ServiceID]services.ServiceAvailabilityChecker, error) {
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

// WrapNetwork implements a networks.NetworkLoader function and wraps the underlying networks.ServiceNetwork with the TestAvalancheNetwork
func (loader TestAvalancheNetworkLoader) WrapNetwork(network *networks.ServiceNetwork) (networks.Network, error) {
	return TestAvalancheNetwork{
		svcNetwork: network,
	}, nil
}
