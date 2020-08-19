package unrequested_chit_spammer_test

import (
	"strconv"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_networks"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_services"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/rpc_workflow_runner"
	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/ids"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigID     networks.ConfigurationID = "normal-config"
	byzantineConfigID      networks.ConfigurationID = "byzantine-config"
	byzantineUsername                               = "byzantine_gecko"
	byzantinePassword                               = "byzant1n3!"
	stakerUsername                                  = "staker_gecko"
	stakerPassword                                  = "test34test!23"
	normalNodeServiceID    networks.ServiceID       = "normal-node"
	byzantineNodePrefix    string                   = "byzantine-node-"
	numberOfByzantineNodes                          = 4
	seedAmount                                      = uint64(50000000000000)
	stakeAmount                                     = uint64(30000000000000)

	networkAcceptanceTimeoutRatio = 0.3
	byzantineBehavior             = "byzantine-behavior"
	chitSpammerBehavior           = "chit-spammer"
)

// StakingNetworkUnrequestedChitSpammerTest tests that a node is able to continue to work normally
// while the network is spammed with chit messages from byzantine peers
type StakingNetworkUnrequestedChitSpammerTest struct {
	ByzantineImageName string
	NormalImageName    string
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkUnrequestedChitSpammerTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))

	// ============= ADD SET OF BYZANTINE NODES AS VALIDATORS ON THE NETWORK ===================
	for i := 0; i < numberOfByzantineNodes; i++ {
		byzClient, err := castedNetwork.GetGeckoClient(networks.ServiceID(byzantineNodePrefix + strconv.Itoa(i)))
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get byzantine client."))
		}
		highLevelByzClient := rpc_workflow_runner.NewRPCWorkFlowRunner(
			byzClient,
			api.UserPass{Username: byzantineUsername, Password: byzantinePassword},
			networkAcceptanceTimeout)
		_, err = highLevelByzClient.ImportGenesisFundsAndStartValidating(seedAmount, stakeAmount)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed add client as a validator."))
		}
		currentStakers, err := byzClient.PChainAPI().GetCurrentValidators(ids.Empty)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
		}
		logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	}

	// =================== ADD NORMAL NODE AS A VALIDATOR ON THE NETWORK =======================
	availabilityChecker, err := castedNetwork.AddService(normalNodeConfigID, normalNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add normal node with high quorum and sample to network."))
	}
	if err = availabilityChecker.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to wait for startup of normal node."))
	}
	normalClient, err := castedNetwork.GetGeckoClient(normalNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get staker client."))
	}
	highLevelNormalClient := rpc_workflow_runner.NewRPCWorkFlowRunner(
		normalClient,
		api.UserPass{Username: stakerUsername, Password: stakerPassword},
		networkAcceptanceTimeout)
	_, err = highLevelNormalClient.ImportGenesisFundsAndStartValidating(seedAmount, stakeAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add client as a validator."))
	}

	// Sleep an additional 10 seconds to ensure that the Validator has time to be added from
	// pending validators to the set of current validators
	time.Sleep(10 * time.Second)
	// ============= VALIDATE NETWORK STATE DESPITE BYZANTINE BEHAVIOR =========================
	currentStakers, err := normalClient.PChainAPI().GetCurrentValidators(ids.Empty)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
	}
	logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	actualNumStakers := len(currentStakers)
	expectedNumStakers := 10
	if actualNumStakers != expectedNumStakers {
		context.AssertTrue(actualNumStakers == expectedNumStakers, stacktrace.NewError("Actual number of stakers, %v, != expected number of stakers, %v", actualNumStakers, expectedNumStakers))
	}
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkUnrequestedChitSpammerTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Define normal node and byzantine node configurations
	serviceConfigs := map[networks.ConfigurationID]ava_networks.TestGeckoNetworkServiceConfig{
		byzantineConfigID: *ava_networks.NewTestGeckoNetworkServiceConfig(
			true,
			ava_services.LOG_LEVEL_DEBUG,
			test.ByzantineImageName,
			2,
			2,
			map[string]string{
				byzantineBehavior: chitSpammerBehavior,
			},
		),
		normalNodeConfigID: *ava_networks.NewTestGeckoNetworkServiceConfig(
			true,
			ava_services.LOG_LEVEL_DEBUG,
			test.NormalImageName,
			6,
			8,
			make(map[string]string),
		),
	}

	// Define the map from service->configuration for the network
	serviceIDConfigMap := map[networks.ServiceID]networks.ConfigurationID{}
	for i := 0; i < numberOfByzantineNodes; i++ {
		serviceIDConfigMap[networks.ServiceID(byzantineNodePrefix+strconv.Itoa(i))] = byzantineConfigID
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
		serviceIDConfigMap)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test StakingNetworkUnrequestedChitSpammerTest) GetExecutionTimeout() time.Duration {
	// TODO drop this when the availabilityChecker doesn't have a sleep, because we spin up a *bunch* of byzantine
	// nodes during test execution
	return 10 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test StakingNetworkUnrequestedChitSpammerTest) GetSetupBuffer() time.Duration {
	return 4 * time.Minute
}
