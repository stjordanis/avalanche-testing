package rpc_workflow_test

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
	stakerUsername    = "staker"
	stakerPassword    = "test34test!23"
	delegatorUsername = "delegator"
	delegatorPassword = "test34test!23"
	seedAmount        = int64(50000000000000)
	stakeAmount       = int64(30000000000000)
	delegatorAmount   = int64(30000000000000)

	regularNodeServiceId   networks.ServiceID = "validator-node"
	delegatorNodeServiceId networks.ServiceID = "delegator-node"

	networkAcceptanceTimeoutRatio = 0.3
	normalNodeConfigId networks.ConfigurationID = 0
)

type StakingNetworkRpcWorkflowTest struct {
	ImageName string
}

func (test StakingNetworkRpcWorkflowTest) Run(network networks.Network, context testsuite.TestContext) {
	// =============================== SETUP GECKO CLIENTS ======================================
	castedNetwork := network.(ava_networks.TestGeckoNetwork)
	stakerClient, err := castedNetwork.GetGeckoClient(regularNodeServiceId)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))

	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get staker client"))
	}
	delegatorClient, err := castedNetwork.GetGeckoClient(delegatorNodeServiceId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get delegator client"))
	}
	stakerNodeId, err := stakerClient.InfoApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get staker node ID."))
	}
	delegatorNodeId, err := delegatorClient.InfoApi().GetNodeId()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get delegator node ID."))
	}
	highLevelStakerClient := rpc_workflow_runner.NewRpcWorkflowRunner(
		stakerClient,
		stakerUsername,
		stakerPassword,
		networkAcceptanceTimeout)
	highLevelDelegatorClient := rpc_workflow_runner.NewRpcWorkflowRunner(
		delegatorClient,
		delegatorUsername,
		delegatorPassword,
		networkAcceptanceTimeout)
  
	// ====================================== ADD VALIDATOR ===============================
	stakerXchainAddress, err := highLevelStakerClient.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not seed XChain account from Genesis."))
	}
	stakerPchainAddress, err := highLevelStakerClient.TransferAvaXChainToPChain(seedAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information"))
	}
	_, err = highLevelDelegatorClient.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not seed XChain account from Genesis."))
	}
	delegatorPchainAddress, err := highLevelDelegatorClient.TransferAvaXChainToPChain(seedAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information"))
	}
	// Adding stakers
	err = highLevelStakerClient.AddValidatorOnSubnet(stakerNodeId, stakerPchainAddress, stakeAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not add staker %s to default subnet.", stakerNodeId))
	}

	// ====================================== VERIFY NETWORK STATE ===============================
	currentStakers, err := stakerClient.PChainApi().GetCurrentValidators(nil)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
	}
	logrus.Debugf("Number of current validators: %d", len(currentStakers))
	actualNumStakers := len(currentStakers)
	expectedNumStakers := 6
	context.AssertTrue(actualNumStakers == expectedNumStakers, stacktrace.NewError("Actual number of stakers, %v, != expected number of stakers, %v", actualNumStakers, expectedNumStakers))

	// ========================= ADD DELEGATOR AND TRANSFER FUNDS TO XCHAIN ======================
	err = highLevelDelegatorClient.AddDelegatorOnSubnet(stakerNodeId, delegatorPchainAddress, delegatorAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not add delegator %s to default subnet.", delegatorNodeId))
	}
	/*
		Currently no way to verify rewards for stakers and delegators because rewards are
		only paid out at the end of the staking period, and the staking period must last at least
		24 hours. This is far too long to be able to test in a CI scenario.
	*/
	remainingStakerAva := seedAmount - stakeAmount
	_, err = highLevelStakerClient.TransferAvaPChainToXChain(stakerPchainAddress, stakerXchainAddress, remainingStakerAva)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to transfer Ava from PChain to XChain."))
	}

	// ================================ VERIFY NETWORK STATE =====================================
	xchainAccountInfo, err := stakerClient.XChainApi().GetBalance(stakerXchainAddress, rpc_workflow_runner.AVA_ASSET_ID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get account info for account %v.", stakerXchainAddress))
	}
	actualRemainingAva := xchainAccountInfo.Balance
	expectedRemainingAva := strconv.FormatInt(remainingStakerAva, 10)
	context.AssertTrue(actualRemainingAva == expectedRemainingAva, stacktrace.NewError("Actual remaining Ava, %v, != expected remaining Ava, %v", actualRemainingAva, expectedRemainingAva))
}

func (test StakingNetworkRpcWorkflowTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Define possible service configurations.
	serviceConfigs := map[networks.ConfigurationID]ava_networks.TestGeckoNetworkServiceConfig{
		normalNodeConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(true, ava_services.LOG_LEVEL_DEBUG, test.ImageName, 2, 2, nil),
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
