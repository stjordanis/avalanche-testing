package unrequested_chit_spammer_test

import (
	"strconv"
	"time"

	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite/rpc_workflow_runner"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigId networks.ConfigurationID = "normal-config"
	byzantineConfigId networks.ConfigurationID = "byzantine-config"
	byzantineUsername = "byzantine_gecko"
	byzantinePassword = "byzant1n3!"
	stakerUsername = "staker_gecko"
	stakerPassword = "test34test!23"
	normalNodeServiceId networks.ServiceID = "normal-node"
	byzantineNodePrefix string = "byzantine-node-"
	numberOfByzantineNodes = 4
	seedAmount               = int64(50000000000000)
	stakeAmount              = int64(30000000000000)

	networkAcceptanceTimeoutRatio = 0.3
)

// ================ Byzantine Test - Spamming Unrequested Chit Messages ===================================
type StakingNetworkUnrequestedChitSpammerTest struct {
	ByzantineImageName string
	NormalImageName    string
}

func (test StakingNetworkUnrequestedChitSpammerTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))

	// ============= ADD SET OF BYZANTINE NODES AS VALIDATORS ON THE NETWORK ===================
	for i := 0; i < numberOfByzantineNodes; i++ {
		byzClient, err := castedNetwork.GetGeckoClient(networks.ServiceID(byzantineNodePrefix + strconv.Itoa(i)))
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get byzantine client."))
		}
		highLevelByzClient := rpc_workflow_runner.NewRpcWorkflowRunner(
			byzClient,
			byzantineUsername,
			byzantinePassword,
			networkAcceptanceTimeout)
		err = highLevelByzClient.GetFundsAndStartValidating(seedAmount, stakeAmount)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed add client as a validator."))
		}
		currentStakers, err := byzClient.PChainApi().GetCurrentValidators(nil)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
		}
		logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	}

	// =================== ADD NORMAL NODE AS A VALIDATOR ON THE NETWORK =======================
	availabilityChecker, err := castedNetwork.AddService(normalNodeConfigId, normalNodeServiceId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add normal node with high quorum and sample to network."))
	}
	if err = availabilityChecker.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to wait for startup of normal node."))
	}
	normalClient, err := castedNetwork.GetGeckoClient(normalNodeServiceId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get staker client."))
	}
	highLevelNormalClient := rpc_workflow_runner.NewRpcWorkflowRunner(
		normalClient,
		stakerUsername,
		stakerPassword,
		networkAcceptanceTimeout)
	err = highLevelNormalClient.GetFundsAndStartValidating(seedAmount, stakeAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed add client as a validator."))
	}

	// ============= VALIDATE NETWORK STATE DESPITE BYZANTINE BEHAVIOR =========================
	currentStakers, err := normalClient.PChainApi().GetCurrentValidators(nil)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
	}
	logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	actualNumStakers := len(currentStakers)
	expectedNumStakers := 10
	context.AssertTrue(actualNumStakers == expectedNumStakers, stacktrace.NewError("Actual number of stakers, %v, != expected number of stakers, %v", actualNumStakers, expectedNumStakers))

}

func (test StakingNetworkUnrequestedChitSpammerTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Define normal node and byzantine node configurations
	serviceConfigs := map[networks.ConfigurationID]ava_networks.TestGeckoNetworkServiceConfig{
		byzantineConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(true,
			ava_services.LOG_LEVEL_DEBUG,
			test.ByzantineImageName,
			2,
			2,
			map[string]string{
				"byzantine-behavior": "chit-spammer",
			},
		),
		normalNodeConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(true,
			ava_services.LOG_LEVEL_DEBUG,
			test.NormalImageName,
			6,
			8,
			make(map[string]string),
		),
	}
	// Define the map from service->configuration for the network
	serviceIdConfigMap := map[networks.ServiceID]networks.ConfigurationID{}
	for i := 0; i < numberOfByzantineNodes; i++ {
		serviceIdConfigMap[networks.ServiceID(byzantineNodePrefix+strconv.Itoa(i))] = byzantineConfigId
	}
	logrus.Debugf("Byzantine Image Name: %s", test.ByzantineImageName)
	logrus.Debugf("Normal Image Name: %s", test.NormalImageName)

	return ava_networks.NewTestGeckoNetworkLoader(
		true,
		test.NormalImageName,
		ava_services.LOG_LEVEL_DEBUG,
		2,
		2,
		serviceConfigs,
		serviceIdConfigMap)
}

func (test StakingNetworkUnrequestedChitSpammerTest) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

func (test StakingNetworkUnrequestedChitSpammerTest) GetSetupBuffer() time.Duration {
	// TODO drop this when the availabilityChecker doesn't have a sleep, because we spin up a *bunch* of nodes before test
	//  execution starts
	return 12 * time.Minute
}
