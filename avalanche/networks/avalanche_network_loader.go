package networks

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	avalancheService "github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/avalanche/services/certs"
	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/services"
	"github.com/palantir/stacktrace"
)

// ========================================================================================================
//                                Avalanche Test Network Loader
// ========================================================================================================

// AvalancheNetworkLoader implements Kurtosis' NetworkLoader interface that's used for creating the test network
// of Avalanche services
type AvalancheNetworkLoader struct {
	// Whether the nodes that get added to the network (boot node and otherwise) will have staking enabled
	isStaking bool

	// The fixed transaction fee for the network
	txFee uint64

	// A registry of the service configurations available for use in this network
	serviceConfigs map[networks.ConfigurationID]TestAvalancheNetworkServiceConfig

	// A mapping of (service ID) -> (service config ID) for the services that the network will initialize with
	initialServiceConfigs map[networks.ServiceID]networks.ConfigurationID

	// service config to be used for each of the five bootstrap nodes
	bootstrapperServiceConfigs []TestAvalancheNetworkServiceConfig
}

// NewDefaultAvalancheNetworkLoader creates a new loader to create a TestAvalancheNetwork with the specified parameters, transparently handling the creation
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
// 	initialServiceConfigs: A map of service_id -> config_id, one per node, that this network will initialize with
func NewDefaultAvalancheNetworkLoader(
	isStaking bool,
	txFee uint64,
	bootstrapNodeServiceConfig TestAvalancheNetworkServiceConfig,
	serviceConfigs map[networks.ConfigurationID]TestAvalancheNetworkServiceConfig,
	initialServiceConfigs map[networks.ServiceID]networks.ConfigurationID) (*AvalancheNetworkLoader, error) {
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
	initialServiceConfigsCopy := make(map[networks.ServiceID]networks.ConfigurationID)
	for serviceID, configID := range initialServiceConfigs {
		if strings.HasPrefix(string(serviceID), bootNodeServiceIDPrefix) {
			return nil, stacktrace.NewError("Service ID %v cannot be used because prefix %v is reserved for boot node services. Choose a service id that does not begin with %v.",
				serviceID,
				bootNodeServiceIDPrefix,
				bootNodeServiceIDPrefix)
		}
		initialServiceConfigsCopy[serviceID] = configID
	}

	// Copy the bootstrap config for each bootstrap node
	bootstrapConfigs := make([]TestAvalancheNetworkServiceConfig, 0, 5)
	for i := 0; i < 5; i++ {
		bootstrapConfigs = append(bootstrapConfigs, bootstrapNodeServiceConfig)
	}

	return &AvalancheNetworkLoader{
		isStaking:                  isStaking,
		serviceConfigs:             serviceConfigsCopy,
		initialServiceConfigs:      initialServiceConfigsCopy,
		txFee:                      txFee,
		bootstrapperServiceConfigs: bootstrapConfigs,
	}, nil
}

func NewCustomBootstrapsAvalancheNetworkLoader(
	isStaking bool,
	txFee uint64,
	bootstrapNodeServiceConfigs []TestAvalancheNetworkServiceConfig,
	serviceConfigs map[networks.ConfigurationID]TestAvalancheNetworkServiceConfig,
	initialServiceConfigs map[networks.ServiceID]networks.ConfigurationID) (*AvalancheNetworkLoader, error) {
	if len(bootstrapNodeServiceConfigs) != 5 {
		return nil, fmt.Errorf("custom bootstraps avalanche network loader requires 5 configs, but received %d", len(bootstrapNodeServiceConfigs))
	}
	bootstrapConfigs := make([]TestAvalancheNetworkServiceConfig, 5)
	copy(bootstrapConfigs, bootstrapNodeServiceConfigs)

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
	initialServiceConfigsCopy := make(map[networks.ServiceID]networks.ConfigurationID)
	for serviceID, configID := range initialServiceConfigs {
		if strings.HasPrefix(string(serviceID), bootNodeServiceIDPrefix) {
			return nil, stacktrace.NewError("Service ID %v cannot be used because prefix %v is reserved for boot node services. Choose a service id that does not begin with %v.",
				serviceID,
				bootNodeServiceIDPrefix,
				bootNodeServiceIDPrefix)
		}
		initialServiceConfigsCopy[serviceID] = configID
	}

	return &AvalancheNetworkLoader{
		isStaking:                  isStaking,
		serviceConfigs:             serviceConfigsCopy,
		initialServiceConfigs:      initialServiceConfigsCopy,
		txFee:                      txFee,
		bootstrapperServiceConfigs: bootstrapConfigs,
	}, nil
}

// ConfigureNetwork defines the network's service configurations to be used
func (loader AvalancheNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
	localNetGenesisStakers := DefaultLocalNetGenesisConfig.Stakers
	bootNodeIDs := make([]string, 0, len(localNetGenesisStakers))
	for _, staker := range DefaultLocalNetGenesisConfig.Stakers {
		bootNodeIDs = append(bootNodeIDs, staker.NodeID)
	}

	// Add boot node configs (five stakers that are part of the local network genesis)
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		configID := networks.ConfigurationID(bootNodeConfigIDPrefix + strconv.Itoa(i))

		certString := localNetGenesisStakers[i].TLSCert
		keyString := localNetGenesisStakers[i].PrivateKey

		certBytes := bytes.NewBufferString(certString)
		keyBytes := bytes.NewBufferString(keyString)

		bootstrapConfig := loader.bootstrapperServiceConfigs[i]

		initializerCore := avalancheService.NewAvalancheServiceInitializerCore(
			bootstrapConfig.snowSampleSize,
			bootstrapConfig.snowQuorumSize,
			loader.txFee,
			loader.isStaking,
			bootstrapConfig.networkInitialTimeout,
			bootstrapConfig.additionalCLIArgs,
			bootNodeIDs[0:i], // Only the node IDs of the already-started nodes
			certs.NewStaticAvalancheCertProvider(*keyBytes, *certBytes),
			bootstrapConfig.serviceLogLevel,
		)
		availabilityCheckerCore := avalancheService.AvalancheServiceAvailabilityCheckerCore{}

		if err := builder.AddConfiguration(configID, bootstrapConfig.imageName, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding bootstrapper node with config ID %v", configID)
		}
	}

	// Add user defined configs
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
func (loader AvalancheNetworkLoader) InitializeNetwork(network *networks.ServiceNetwork) (map[networks.ServiceID]services.ServiceAvailabilityChecker, error) {
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
	for serviceID, configID := range loader.initialServiceConfigs {
		checker, err := network.AddService(configID, serviceID, bootstrapperServiceIDs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding non-boot node with ID %v and config ID %v", serviceID, configID)
		}
		availabilityCheckers[serviceID] = *checker
	}
	return availabilityCheckers, nil
}

// WrapNetwork implements a networks.NetworkLoader function and wraps the underlying networks.ServiceNetwork with the TestAvalancheNetwork
func (loader AvalancheNetworkLoader) WrapNetwork(network *networks.ServiceNetwork) (networks.Network, error) {
	return TestAvalancheNetwork{
		svcNetwork: network,
	}, nil
}
