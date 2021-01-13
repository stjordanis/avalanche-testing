package networks

import (
	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/services"

	"strconv"

	avalancheService "github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/utils/constants"

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
	containerStopTimeoutSeconds = 30
)

// TestAvalancheNetwork wraps Kurtosis' ServiceNetwork that is meant to be the interface tests use for interacting with Avalanche
// networks
type TestAvalancheNetwork struct {
	networks.Network

	svcNetwork *networks.ServiceNetwork
}

// GetAvalancheClient returns the API Client for the node with the given service ID
func (network TestAvalancheNetwork) GetAvalancheClient(serviceID networks.ServiceID) (*avalancheService.Client, error) {
	node, err := network.svcNetwork.GetService(serviceID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving service node with ID %v", serviceID)
	}
	service := node.Service.(avalancheService.AvalancheService)
	jsonRPCSocket := service.GetJSONRPCSocket()
	return avalancheService.NewClient(jsonRPCSocket.GetIPAddr(), jsonRPCSocket.GetPort(), constants.DefaultRequestTimeout)
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
// 		An availability checker that will return true when the newly-added service is available
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
	if err := network.svcNetwork.RemoveService(serviceID, containerStopTimeoutSeconds); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service with ID %v", serviceID)
	}
	return nil
}
