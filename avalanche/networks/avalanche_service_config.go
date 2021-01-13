package networks

import (
	"time"

	avalancheService "github.com/ava-labs/avalanche-testing/avalanche/services"
)

const (
	// The byzantine behavior CLI parameter for configuring byzantine nodes
	byzantineBehaviorKey = "byzantine-behavior"
)

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

	// TODO include these params
	// epochFirstTransitionTime time.Time
	// epochDuration            time.Duration

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
	additionalCLIArgs map[string]string,
) *TestAvalancheNetworkServiceConfig {
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

// NewDefaultAvalancheNetworkServiceConfig returns a default service config
// using [imageName]
func NewDefaultAvalancheNetworkServiceConfig(imageName string) *TestAvalancheNetworkServiceConfig {
	return &TestAvalancheNetworkServiceConfig{
		varyCerts:             true,
		serviceLogLevel:       avalancheService.DEBUG,
		imageName:             imageName,
		snowQuorumSize:        2,
		snowSampleSize:        2,
		networkInitialTimeout: 2 * time.Second,
		additionalCLIArgs:     make(map[string]string),
	}
}

// NewByzantineServiceConfig returns a service config
// using [imageName] as the byzantine image and [byzantineBehavior]
// as the byzantine behavior for the node
func NewAvalancheByzantineServiceConfig(imageName, byzantineBehavior string) *TestAvalancheNetworkServiceConfig {
	return &TestAvalancheNetworkServiceConfig{
		varyCerts:             true,
		serviceLogLevel:       avalancheService.DEBUG,
		imageName:             imageName,
		snowQuorumSize:        2,
		snowSampleSize:        2,
		networkInitialTimeout: 2 * time.Second,
		additionalCLIArgs:     map[string]string{byzantineBehaviorKey: byzantineBehavior},
	}
}

// SetCLIArgs replaces the existing [additionalCLIArgs] with [args]
func (config *TestAvalancheNetworkServiceConfig) SetCLIArgs(args map[string]string) {
	config.additionalCLIArgs = args
}
