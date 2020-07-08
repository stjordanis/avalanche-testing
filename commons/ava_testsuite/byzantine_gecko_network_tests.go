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
	BYZANTINE_SERVICE_ID = 2
	BYZANTINE_USERNAME = "byzantine_gecko"
	BYZANTINE_PASSWORD = "byzant1n3!"
)
// ================ Byzantine Test - Spamming Unrequested Chit Messages ===================================
type StakingNetworkUnrequestedChitSpammerTest struct{
	unrequestedChitSpammerImageName *string
}
func (test StakingNetworkUnrequestedChitSpammerTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	/*for i := 0; i < 5; i++ {
		_, err := addServiceIdAsValidator(
			castedNetwork,
			i,
			BYZANTINE_USERNAME,
			BYZANTINE_PASSWORD,
			SEED_AMOUNT,
			STAKE_AMOUNT)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get high level byzantine client as a validator."))
		}
	}*/
	highLevelNormalClient, err := addServiceIdAsValidator(
		castedNetwork,
		4,
		STAKER_USERNAME,
		STAKER_PASSWORD,
		SEED_AMOUNT,
		STAKE_AMOUNT)
	normalClient := highLevelNormalClient.GetLowLevelClient()
	currentStakers, err := normalClient.PChainApi().GetCurrentValidators(nil)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
	}
	logrus.Debugf("Number of current stakers: %d", len(currentStakers))
	actualNumStakers := len(currentStakers)
	expectedNumStakers := 6
	context.AssertTrue(actualNumStakers == expectedNumStakers, stacktrace.NewError("Actual number of stakers, %v, != expected number of stakers, %v", actualNumStakers, expectedNumStakers))
}
func (test StakingNetworkUnrequestedChitSpammerTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return getByzantineNetworkLoader(map[int]int{
		/*0:           NORMAL_NODE_CONFIG_ID,
		1:           NORMAL_NODE_CONFIG_ID,
		2:           NORMAL_NODE_CONFIG_ID,
		3: NORMAL_NODE_CONFIG_ID,*/
		4: NORMAL_NODE_CONFIG_ID,
	}, test.unrequestedChitSpammerImageName)
}
func (test StakingNetworkUnrequestedChitSpammerTest) GetTimeout() time.Duration {
	return 90 * time.Second
}



// =============== Helper functions =============================

/*
Args:
	desiredServices: Mapping of service_id -> configuration_id for all services *in addition to the boot nodes* that the user wants
*/
func getByzantineNetworkLoader(desiredServices map[int]int, byzantineImageName *string) (testsuite.TestNetworkLoader, error) {
	serviceConfigs := map[int]ava_networks.TestGeckoNetworkServiceConfig{
		NORMAL_NODE_CONFIG_ID: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, nil),
		BYZANTINE_CONFIG_ID:   *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, byzantineImageName),
	}
	return ava_networks.NewTestGeckoNetworkLoader(
		ava_services.LOG_LEVEL_DEBUG,
		true,
		serviceConfigs,
		desiredServices,
		4,
		6)
}

func addServiceIdAsValidator(
		network ava_networks.TestGeckoNetwork,
		serviceId int,
		username string,
		password string,
		seedAmount int64,
		stakeAmount int64) (*ava_networks.HighLevelGeckoClient, error) {
	client, err := network.GetGeckoClient(serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get byzantine client.")
	}
	highLevelClient := ava_networks.NewHighLevelGeckoClient(
		client,
		username,
		password)
	err = highLevelClient.GetFundsAndStartValidating(seedAmount, stakeAmount)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed add client as a validator.")
	}
	return highLevelClient, nil
}
