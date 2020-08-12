package rpc_workflow_test

import (
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_networks"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
)

const (
	stakerUsername    = "staker"
	stakerPassword    = "test34test!23"
	delegatorUsername = "delegator"
	delegatorPassword = "test34test!23"
	seedAmount        = uint64(50000000000000)
	stakeAmount       = uint64(30000000000000)
	delegatorAmount   = uint64(30000000000000)

	regularNodeServiceId   networks.ServiceID = "validator-node"
	delegatorNodeServiceId networks.ServiceID = "delegator-node"

	networkAcceptanceTimeoutRatio                          = 0.3
	normalNodeConfigId            networks.ConfigurationID = "normal-config"
)

type StakingNetworkRpcWorkflowTest struct {
	ImageName string
}

func (test StakingNetworkRpcWorkflowTest) Run(network networks.Network, context testsuite.TestContext) {
	// =============================== SETUP GECKO CLIENTS ======================================
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))
	stakerClient, err := castedNetwork.GetGeckoClient(regularNodeServiceId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get staker client"))
	}

	delegatorClient, err := castedNetwork.GetGeckoClient(delegatorNodeServiceId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get delegator client"))
	}

	executor := NewRPCWorkflowTestExecutor(stakerClient, delegatorClient, networkAcceptanceTimeout)

	if err := executor.ExecuteTest(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "RPCWorkflow Test failed."))
	}
}

func (test StakingNetworkRpcWorkflowTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Define possible service configurations.
	serviceConfigs := map[networks.ConfigurationID]ava_networks.TestGeckoNetworkServiceConfig{
		normalNodeConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, test.ImageName, 2, 2, make(map[string]string)),
	}
	// Define which services use which configurations.
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{
		regularNodeServiceId:   normalNodeConfigId,
		delegatorNodeServiceId: normalNodeConfigId,
	}
	// Return a Gecko test net with this service:configuration mapping.
	return ava_networks.NewTestGeckoNetworkLoader(
		true,
		test.ImageName,
		ava_services.LOG_LEVEL_DEBUG,
		2,
		2,
		serviceConfigs,
		desiredServices)
}

func (test StakingNetworkRpcWorkflowTest) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

func (test StakingNetworkRpcWorkflowTest) GetSetupBuffer() time.Duration {
	// TODO drop this down when the availability checker doesn't have a sleep (becuase we spin up a bunch of nodes before the test starts executing)
	return 6 * time.Minute
}
