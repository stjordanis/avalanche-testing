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

// NewRPCWorkflowTestExecutor ...
func NewRPCWorkflowTestExecutor(stakerClient, delegatorClient *apis.Client, acceptanceTimeout time.Duration) ava_testsuite.AvalancheTester {
	return &executor{
		stakerClient:      stakerClient,
		delegatorClient:   delegatorClient,
		acceptanceTimeout: acceptanceTimeout,
	}
}

// ExecuteTest ...
func (e *executor) ExecuteTest() error {
	stakerNodeID, err := e.stakerClient.InfoAPI().GetNodeID()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get staker node ID.")
	}
	delegatorNodeID, err := e.delegatorClient.InfoAPI().GetNodeID()
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

	// ====================================== SEED ACCOUNTS ===============================
	stakerXchainAddress, err := highLevelStakerClient.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not seed XChain account from Genesis.")
	}
	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXchainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain balance for staker.")
	}
	stakerPchainAddress, err := highLevelStakerClient.TransferAvaXChainToPChain(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information")
	}
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPchainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain balance after X -> P Transfer.")
	}
	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXchainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "X Chain Balance not updated correctly after X -> P Transfer for validator")
	}

	time.Sleep(5 * time.Second)
	delegatorXChainAddress, err := highLevelDelegatorClient.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not seed X Chain account from Genesis.")
	}
	if err := highLevelDelegatorClient.VerifyXChainAVABalance(delegatorXChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance after seeding account.")
	}
	delegatorPchainAddress, err := highLevelDelegatorClient.TransferAvaXChainToPChain(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVA from X Chain to P Chain account.")
	}
	if err := highLevelDelegatorClient.VerifyPChainBalance(delegatorPchainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain balance after X -> P Transfer for Delegator.")
	}
	if err := highLevelDelegatorClient.VerifyXChainAVABalance(delegatorXChainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance after X -> P Transfer for Delegator")
	}

	// Adds new staker to the network and block until its staking period begins
	err = highLevelStakerClient.AddValidatorOnSubnet(stakerNodeID, stakerPchainAddress, stakeAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add staker %s to default subnet.", stakerNodeID)
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
	expectedStakerBalance := seedAmount - stakeAmount
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPchainAddress, expectedStakerBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after adding default subnet validator to the network")
	}

	// ====================================== ADD DELEGATOR ======================================
	err = highLevelDelegatorClient.AddDelegatorOnSubnet(stakerNodeID, delegatorPchainAddress, delegatorAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add delegator %s to default subnet.", delegatorNodeID)
	}
	expectedDelegatorBalance := seedAmount - delegatorAmount
	if err := highLevelDelegatorClient.VerifyPChainBalance(delegatorPchainAddress, expectedDelegatorBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after adding a new delegator to the network.")
	}

	// Transfer funds back to the X Chain
	_, err = highLevelStakerClient.TransferAvaPChainToXChain(stakerPchainAddress, stakerXchainAddress, expectedStakerBalance)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to transfer Ava from PChain to XChain.")
	}
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPchainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after P -> X Transfer.")
	}
	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXchainAddress, expectedStakerBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer.")
	}

	return nil
}
