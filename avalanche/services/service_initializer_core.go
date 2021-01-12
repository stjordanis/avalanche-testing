package services

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/services"

	"github.com/ava-labs/avalanche-testing/avalanche/services/certs"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	httpPort    = 9650
	stakingPort = 9651

	stakingTLSCertFileID = "staking-tls-cert"
	stakingTLSKeyFileID  = "staking-tls-key"

	testVolumeMountpoint = "/shared"
	avalancheBinary      = "/avalanchego/build/avalanchego"
)

// AvalancheLogLevel specifies the log level for an Avalanche client
type AvalancheLogLevel string

// Log levels
const (
	VERBOSE AvalancheLogLevel = "verbo"
	DEBUG   AvalancheLogLevel = "debug"
	INFO    AvalancheLogLevel = "info"
)

// AvalancheServiceInitializerCore implements Kurtosis' services.ServiceInitializerCore used to initialize an Avalanche service
type AvalancheServiceInitializerCore struct {
	// Snow protocol sample size
	snowSampleSize int

	// Snow protocol quorum size
	snowQuorumSize int

	// Whether the node should be started with staking enabled
	stakingEnabled bool

	// The fixed transaction fee for the network
	txFee uint64

	// The initial timeout for the network
	networkInitialTimeout time.Duration

	// TODO Switch these to be named properties of this struct, so that we're being explicit about what arguments
	//  are consumed
	// A set of CLI args that will be passed as-is to the Avalanche service
	additionalCLIArgs map[string]string

	// The node IDs of the nodes this node should bootstrap from
	bootstrapperNodeIDs []string

	// Cert provider that should be used when initializing the Avalanche service
	certProvider certs.AvalancheCertProvider

	// Log level that the Avalanche service should start with
	logLevel AvalancheLogLevel
}

// NewAvalancheServiceInitializerCore creates a new Avalanche service initializer core with the following parameters:
// Args:
// 		snowSampleSize: Sample size for Snow consensus protocol
// 		snowQuroumSize: Quorum size for Snow consensus protocol
// 		stakingEnabled: Whether this node will use staking
// 		cliArgs: A mapping of cli_arg -> cli_arg_value that will be passed as-is to the Avalanche node
// 		bootstrapperNodeIDs: The node IDs of the bootstrapper nodes that this node will connect to. While this *seems* unintuitive
// 			why this would be required, it's because Avalanche doesn't actually use certs. So, to prevent against man-in-the-middle attacks,
// 			the user is required to manually specify the node IDs of the nodese it's connecting to.
// 		certProvider: Provides the certs used by the Avalanche services generated by this core
// 		logLevel: The loglevel that the Avalanche node should output at.
// Returns:
// 		An intializer core for creating Avalanche nodes with the specified parameers.
func NewAvalancheServiceInitializerCore(
	snowSampleSize int,
	snowQuorumSize int,
	txFee uint64,
	stakingEnabled bool,
	networkInitialTimeout time.Duration,
	additionalCLIArgs map[string]string,
	bootstrapperNodeIDs []string,
	certProvider certs.AvalancheCertProvider,
	logLevel AvalancheLogLevel) *AvalancheServiceInitializerCore {
	// Defensive copy
	bootstrapperIDsCopy := make([]string, 0, len(bootstrapperNodeIDs))
	bootstrapperIDsCopy = append(bootstrapperIDsCopy, bootstrapperNodeIDs...)

	return &AvalancheServiceInitializerCore{
		snowSampleSize:        snowSampleSize,
		snowQuorumSize:        snowQuorumSize,
		txFee:                 txFee,
		stakingEnabled:        stakingEnabled,
		networkInitialTimeout: networkInitialTimeout,
		additionalCLIArgs:     additionalCLIArgs,
		bootstrapperNodeIDs:   bootstrapperIDsCopy,
		certProvider:          certProvider,
		logLevel:              logLevel,
	}
}

// GetUsedPorts implements services.ServiceInitializerCore to declare the ports used by the node
func (core AvalancheServiceInitializerCore) GetUsedPorts() map[int]bool {
	return map[int]bool{
		httpPort:    true,
		stakingPort: true,
	}
}

// GetFilesToMount implements services.ServiceInitializerCore to declare the files used by the node
func (core AvalancheServiceInitializerCore) GetFilesToMount() map[string]bool {
	if core.stakingEnabled {
		return map[string]bool{
			stakingTLSCertFileID: true,
			stakingTLSKeyFileID:  true,
		}
	}
	return make(map[string]bool)
}

// InitializeMountedFiles implementats services.ServiceInitializerCore to initialize the file needed by the node
func (core AvalancheServiceInitializerCore) InitializeMountedFiles(osFiles map[string]*os.File, dependencies []services.Service) error {
	certFilePointer := osFiles[stakingTLSCertFileID]
	keyFilePointer := osFiles[stakingTLSKeyFileID]
	certPEM, keyPEM, err := core.certProvider.GetCertAndKey()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get cert & key when initializing service")
	}
	if _, err := certFilePointer.Write(certPEM.Bytes()); err != nil {
		return err
	}
	if _, err := keyFilePointer.Write(keyPEM.Bytes()); err != nil {
		return err
	}
	return nil
}

// GetStartCommand implements services.ServiceInitializerCore to build the command line that will be used to launch an Avalanche node
// The IP placeholder is a string that can be used in place of the IP, since we don't yet know the IP when we ask to start a new service
func (core AvalancheServiceInitializerCore) GetStartCommand(mountedFileFilepaths map[string]string, ipPlaceholder string, dependencies []services.Service) ([]string, error) {
	numBootNodeIDs := len(core.bootstrapperNodeIDs)
	numDependencies := len(dependencies)
	if numDependencies > numBootNodeIDs {
		return nil, stacktrace.NewError(
			"Avalanche service is being started with %v dependencies but only %v boot node IDs have been configured",
			numDependencies,
			numBootNodeIDs,
		)
	}

	publicIPFlag := fmt.Sprintf("--public-ip=%s", ipPlaceholder)
	commandList := []string{
		avalancheBinary,
		publicIPFlag,
		"--network-id=local",
		fmt.Sprintf("--http-port=%d", httpPort),
		"--http-host=", // Leave empty to make API openly accessible
		fmt.Sprintf("--staking-port=%d", stakingPort),
		fmt.Sprintf("--log-level=%s", core.logLevel),
		fmt.Sprintf("--snow-sample-size=%d", core.snowSampleSize),
		fmt.Sprintf("--snow-quorum-size=%d", core.snowQuorumSize),
		fmt.Sprintf("--staking-enabled=%v", core.stakingEnabled),
		fmt.Sprintf("--tx-fee=%d", core.txFee),
		fmt.Sprintf("--network-initial-timeout=%s", core.networkInitialTimeout),
	}

	if core.stakingEnabled {
		certFilepath, found := mountedFileFilepaths[stakingTLSCertFileID]
		if !found {
			return nil, stacktrace.NewError("Could not find file key '%v' in the mounted filepaths map; this is likely a code bug", stakingTLSCertFileID)
		}
		keyFilepath, found := mountedFileFilepaths[stakingTLSKeyFileID]
		if !found {
			return nil, stacktrace.NewError("Could not find file key '%v' in the mounted filepaths map; this is likely a code bug", stakingTLSKeyFileID)
		}
		commandList = append(commandList, fmt.Sprintf("--staking-tls-cert-file=%s", certFilepath))
		commandList = append(commandList, fmt.Sprintf("--staking-tls-key-file=%s", keyFilepath))

		// NOTE: This seems weird, BUT there's a reason for it: An avalanche node doesn't use certs, and instead relies on
		//  the user explicitly passing in the node ID of the bootstrapper it wants. This prevents man-in-the-middle
		//  attacks, just like using a cert would. Us hardcoding this bootstrapper ID here is the equivalent
		//  of a user knowing the node ID in advance, which provides the same level of protection.
		commandList = append(commandList, "--bootstrap-ids="+strings.Join(core.bootstrapperNodeIDs, ","))
	}

	if len(dependencies) > 0 {
		avaDependencies := make([]NodeService, 0, len(dependencies))
		for _, service := range dependencies {
			avaDependencies = append(avaDependencies, service.(NodeService))
		}

		socketStrs := make([]string, 0, len(avaDependencies))
		for _, service := range avaDependencies {
			socket := service.GetStakingSocket()
			socketStrs = append(socketStrs, fmt.Sprintf("%s:%d", socket.GetIPAddr(), socket.GetPort()))
		}
		joinedSockets := strings.Join(socketStrs, ",")
		commandList = append(commandList, "--bootstrap-ips="+joinedSockets)
	}

	// Append additional CLI arguments
	// These are added as is with no additional checking
	for param, argument := range core.additionalCLIArgs {
		commandList = append(commandList, fmt.Sprintf("--%s=%s", param, argument))
	}

	logrus.Debugf("Command list: %s", commandList)
	return commandList, nil
}

// GetServiceFromIp implements services.ServiceInitializerCore function to take the IP address of the Docker container that Kurtosis
// launches the Avalanche node inside and wrap it with our AvalancheService implementation of NodeService
func (core AvalancheServiceInitializerCore) GetServiceFromIp(ipAddr string) services.Service {
	return AvalancheService{
		ipAddr:      ipAddr,
		stakingPort: stakingPort,
		jsonRPCPort: httpPort,
	}
}

// GetTestVolumeMountpoint implements services.ServiceInitializerCore to declare the path on the Avalanche Docker image where the test
// Docker volume should be mounted on
func (core AvalancheServiceInitializerCore) GetTestVolumeMountpoint() string {
	return testVolumeMountpoint
}
