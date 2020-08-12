package rpc_workflow_test

import (
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/rpc_workflow_runner"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/gecko/utils/constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type executor struct {
	stakerClient, delegatorClient *apis.Client
	acceptanceTimeout             time.Duration
}

func NewRPCWorkflowTestExecutor(stakerClient, delegatorClient *apis.Client, acceptanceTimeout time.Duration) ava_testsuite.AvalancheTester {
	return &executor{
		stakerClient:      stakerClient,
		delegatorClient:   delegatorClient,
		acceptanceTimeout: acceptanceTimeout,
	}
}

func (e *executor) ExecuteTest() error {
	stakerNodeId, err := e.stakerClient.InfoAPI().GetNodeID()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get staker node ID.")
	}
	delegatorNodeId, err := e.delegatorClient.InfoAPI().GetNodeID()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get delegator node ID.")
	}
	highLevelStakerClient := rpc_workflow_runner.NewRpcWorkflowRunner(
		e.stakerClient,
		stakerUsername,
		stakerPassword,
		e.acceptanceTimeout,
	)
	highLevelDelegatorClient := rpc_workflow_runner.NewRpcWorkflowRunner(
		e.delegatorClient,
		delegatorUsername,
		delegatorPassword,
		e.acceptanceTimeout,
	)

	// ====================================== ADD VALIDATOR ===============================
	stakerXchainAddress, err := highLevelStakerClient.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not seed XChain account from Genesis.")
	}
	stakerPchainAddress, err := highLevelStakerClient.TransferAvaXChainToPChain(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information")
	}

	time.Sleep(5 * time.Second)
	_, err = highLevelDelegatorClient.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not seed XChain account from Genesis.")
	}

	delegatorPchainAddress, err := highLevelDelegatorClient.TransferAvaXChainToPChain(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information")
	}
	// Adding stakers
	err = highLevelStakerClient.AddValidatorOnSubnet(stakerNodeId, stakerPchainAddress, stakeAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add staker %s to default subnet.", stakerNodeId)
	}

	// ====================================== VERIFY NETWORK STATE ===============================
	currentStakers, err := e.stakerClient.PChainAPI().GetCurrentValidators(constants.DefaultSubnetID)
	if err != nil {
		return stacktrace.Propagate(err, "Could not get current stakers.")
	}
	logrus.Debugf("Number of current validators: %d", len(currentStakers))
	actualNumStakers := len(currentStakers)
	expectedNumStakers := 6
	if actualNumStakers != expectedNumStakers {
		return stacktrace.NewError("Actual number of stakers, %v, != expected number of stakers, %v", actualNumStakers, expectedNumStakers)
	}

	// ========================= ADD DELEGATOR AND TRANSFER FUNDS TO XCHAIN ======================
	err = highLevelDelegatorClient.AddDelegatorOnSubnet(stakerNodeId, delegatorPchainAddress, delegatorAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add delegator %s to default subnet.", delegatorNodeId)
	}
	/*
		Currently no way to verify rewards for stakers and delegators because rewards are
		only paid out at the end of the staking period, and the staking period must last at least
		24 hours. This is far too long to be able to test in a CI scenario.
	*/
	remainingStakerAva := seedAmount - stakeAmount
	_, err = highLevelStakerClient.TransferAvaPChainToXChain(stakerPchainAddress, stakerXchainAddress, remainingStakerAva)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to transfer Ava from PChain to XChain.")
	}

	// ================================ VERIFY NETWORK STATE =====================================
	balanceInfo, err := e.stakerClient.XChainAPI().GetBalance(stakerXchainAddress, rpc_workflow_runner.AVA_ASSET_ID)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get account info for account %v.", stakerXchainAddress)
	}
	actualRemainingAva := uint64(balanceInfo.Balance)
	if actualRemainingAva != remainingStakerAva {
		return stacktrace.NewError("Actual remaining Ava, %v, != expected remaining Ava, %v", actualRemainingAva, remainingStakerAva)
	}

	return nil
}
