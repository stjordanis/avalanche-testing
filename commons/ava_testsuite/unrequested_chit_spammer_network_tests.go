package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	BYZANTINE_CONFIG_ID = 2
	BYZANTINE_USERNAME = "byzantine_gecko"
	BYZANTINE_PASSWORD = "byzant1n3!"
)
// ================ Byzantine Test - Spamming Unrequested Chit Messages ===================================
type StakingNetworkUnrequestedChitSpammerTest struct{
	unrequestedChitSpammerImageName *string
	normalImageName *string
}
func (test StakingNetworkUnrequestedChitSpammerTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	for i := 0; i < 4; i++ {
		byzClient, err := castedNetwork.GetGeckoClient(i)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get byzantine client."))
		}
		highLevelByzClient := ava_networks.NewHighLevelGeckoClient(
			byzClient,
			BYZANTINE_USERNAME,
			BYZANTINE_PASSWORD)
		err = highLevelByzClient.GetFundsAndStartValidating(SEED_AMOUNT, STAKE_AMOUNT)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err,"Failed add client as a validator."))
		}
		currentStakers, err := byzClient.PChainApi().GetCurrentValidators(nil)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
		}
		logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	}
	availabilityChecker, err := castedNetwork.AddService(NORMAL_NODE_CONFIG_ID, 4)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add normal node with high quorum and sample to network."))
	}
	err = availabilityChecker.WaitForStartup()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to wait for startup of normal node."))
	}
	normalClient, err := castedNetwork.GetGeckoClient(4)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err,"Failed to get staker client."))
	}
	highLevelNormalClient := ava_networks.NewHighLevelGeckoClient(
		normalClient,
		STAKER_USERNAME,
		STAKER_PASSWORD)
	err = highLevelNormalClient.GetFundsAndStartValidating(SEED_AMOUNT, STAKE_AMOUNT)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err,"Failed add client as a validator."))
	}
	currentStakers, err := normalClient.PChainApi().GetCurrentValidators(nil)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
	}
	logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	actualNumStakers := len(currentStakers)
	expectedNumStakers := 10
	context.AssertTrue(actualNumStakers == expectedNumStakers, stacktrace.NewError("Actual number of stakers, %v, != expected number of stakers, %v", actualNumStakers, expectedNumStakers))
}
func (test StakingNetworkUnrequestedChitSpammerTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getByzantineNetworkLoader(map[int]int{
		0:           BYZANTINE_CONFIG_ID,
		1:           BYZANTINE_CONFIG_ID,
		2:           BYZANTINE_CONFIG_ID,
		3:           BYZANTINE_CONFIG_ID,
	}, test.unrequestedChitSpammerImageName, test.normalImageName)
}
func (test StakingNetworkUnrequestedChitSpammerTest) GetTimeout() time.Duration {
	return 720 * time.Second
}



// =============== Helper functions =============================

/*
Args:
	desiredServices: Mapping of service_id -> configuration_id for all services *in addition to the boot nodes* that the user wants
*/
func getByzantineNetworkLoader(desiredServices map[int]int, byzantineImageName *string, normalImageName *string) (testsuite.TestNetworkLoader, error) {
	serviceConfigs := map[int]ava_networks.TestGeckoNetworkServiceConfig{
		NORMAL_NODE_CONFIG_ID: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, normalImageName, 6, 8),
		BYZANTINE_CONFIG_ID:   *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, byzantineImageName, 2, 2),
	}
	logrus.Debugf("Byzantine Image Name: %s", *byzantineImageName)
	logrus.Debugf("Normal Image Name: %s", *normalImageName)
	return ava_networks.NewTestGeckoNetworkLoader(
		ava_services.LOG_LEVEL_DEBUG,
		true,
		serviceConfigs,
		desiredServices,
		2,
		2)
}
