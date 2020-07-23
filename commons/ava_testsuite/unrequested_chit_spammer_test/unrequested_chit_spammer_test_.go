package unrequested_chit_spammer_test

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	normalNodeConfigId networks.ConfigurationID = 1
	byzantineConfigId networks.ConfigurationID = 2
	byzantineUsername = "byzantine_gecko"
	byzantinePassword = "byzant1n3!"
	stakerUsername = "staker_gecko"
	stakerPassword = "test34test!23"
	normalNodeServiceId networks.ServiceID = 4
	seedAmount               = int64(50000000000000)
	stakeAmount              = int64(30000000000000)

	networkAcceptanceTimeoutRatio = 0.3
)
// ================ Byzantine Test - Spamming Unrequested Chit Messages ===================================
type StakingNetworkUnrequestedChitSpammerTest struct{
	UnrequestedChitSpammerImageName string
	NormalImageName                 string
}
func (test StakingNetworkUnrequestedChitSpammerTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	networkAcceptanceTimeout := time.Duration(int(networkAcceptanceTimeoutRatio * test.GetTimeout().Seconds()))

	for i := 0; i < int(normalNodeServiceId); i++ {
		byzClient, err := castedNetwork.GetGeckoClient(networks.ServiceID(i))
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get byzantine client."))
		}
		highLevelByzClient := ava_networks.NewHighLevelGeckoClient(
			byzClient,
			byzantineUsername,
			byzantinePassword,
			networkAcceptanceTimeout)
		err = highLevelByzClient.GetFundsAndStartValidating(seedAmount, stakeAmount)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err,"Failed add client as a validator."))
		}
		currentStakers, err := byzClient.PChainApi().GetCurrentValidators(nil)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
		}
		logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	}
	availabilityChecker, err := castedNetwork.AddService(normalNodeConfigId, normalNodeServiceId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add normal node with high quorum and sample to network."))
	}
	if err = availabilityChecker.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to wait for startup of normal node."))
	}
	normalClient, err := castedNetwork.GetGeckoClient(normalNodeServiceId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err,"Failed to get staker client."))
	}
	highLevelNormalClient := ava_networks.NewHighLevelGeckoClient(
		normalClient,
		stakerUsername,
		stakerPassword,
		networkAcceptanceTimeout)
	err = highLevelNormalClient.GetFundsAndStartValidating(seedAmount, stakeAmount)
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
func (test StakingNetworkUnrequestedChitSpammerTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	serviceIdConfigMap := map[networks.ServiceID]networks.ConfigurationID{}
	for i := 0; i < int(normalNodeServiceId); i++ {
		serviceIdConfigMap[networks.ServiceID(i)] = byzantineConfigId
	}
	return getByzantineNetworkLoader(serviceIdConfigMap, test.UnrequestedChitSpammerImageName, test.NormalImageName)
}
func (test StakingNetworkUnrequestedChitSpammerTest) GetTimeout() time.Duration {
	return 720 * time.Second
}



// =============== Helper functions =============================

/*
Args:
	desiredServices: Mapping of service_id -> configuration_id for all services *in addition to the boot nodes* that the user wants
*/
func getByzantineNetworkLoader(
			desiredServices map[networks.ServiceID]networks.ConfigurationID,
			byzantineImageName string,
			normalImageName string) (networks.NetworkLoader, error) {
	serviceConfigs := map[networks.ConfigurationID]ava_networks.TestGeckoNetworkServiceConfig{
		normalNodeConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, normalImageName, 6, 8),
		byzantineConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, byzantineImageName, 2, 2),
	}
	logrus.Debugf("Byzantine Image Name: %s", byzantineImageName)
	logrus.Debugf("Normal Image Name: %s", normalImageName)

	return ava_networks.NewTestGeckoNetworkLoader(
		true,
		normalImageName,
		ava_services.LOG_LEVEL_DEBUG,
		2,
		2,
		serviceConfigs,
		desiredServices)
}
