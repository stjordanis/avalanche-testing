package workflow

import (
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	avalancheNetwork "github.com/ava-labs/avalanche-testing/avalanche/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	regularNodeServiceID   networks.ServiceID = "validator-node"
	delegatorNodeServiceID networks.ServiceID = "delegator-node"

	networkAcceptanceTimeoutRatio                          = 0.3
	normalNodeConfigID            networks.ConfigurationID = "normal-config"
)

// StakingNetworkRPCWorkflowTest ...
type StakingNetworkRPCWorkflowTest struct {
	ImageName string
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkRPCWorkflowTest) Run(network networks.Network, context testsuite.TestContext) {
	// =============================== SETUP AVALANCHE CLIENTS ======================================
	castedNetwork := network.(avalancheNetwork.TestAvalancheNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))
	stakerClient, err := castedNetwork.GetAvalancheClient(regularNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get staker client"))
	}

	delegatorClient, err := castedNetwork.GetAvalancheClient(delegatorNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get delegator client"))
	}

	executor := NewRPCWorkflowTestExecutor(stakerClient, delegatorClient, networkAcceptanceTimeout)

	logrus.Infof("Set up RPCWorkFlowTest. Executing...")
	if err := executor.ExecuteTest(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "RPCWorkflow Test failed."))
	}
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkRPCWorkflowTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Define possible service configurations.
	normalServiceConfig := *avalancheNetwork.NewDefaultAvalancheNetworkServiceConfig(test.ImageName)
	serviceConfigs := map[networks.ConfigurationID]avalancheNetwork.TestAvalancheNetworkServiceConfig{
		normalNodeConfigID: normalServiceConfig,
	}
	// Define which services use which configurations.
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{
		regularNodeServiceID:   normalNodeConfigID,
		delegatorNodeServiceID: normalNodeConfigID,
	}
	// Return an Avalanche Test Network with this service:configuration mapping.
	return avalancheNetwork.NewTestAvalancheNetworkLoader(
		true,
		0,
		normalServiceConfig,
		serviceConfigs,
		desiredServices,
	)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test StakingNetworkRPCWorkflowTest) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test StakingNetworkRPCWorkflowTest) GetSetupBuffer() time.Duration {
	// TODO drop this down when the availability checker doesn't have a sleep (because we spin up a bunch of nodes before the test starts executing)
	return 6 * time.Minute
}
