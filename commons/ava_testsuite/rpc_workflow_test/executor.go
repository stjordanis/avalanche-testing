package rpc_workflow_test

import (
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/rpc_workflow_runner"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/utils/constants"
	"github.com/ava-labs/gecko/utils/units"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	genesisUsername   = "genesis"
	genesisPassword   = "MyNameIs!Jeff"
	stakerUsername    = "staker"
	stakerPassword    = "test34test!23"
	delegatorUsername = "delegator"
	delegatorPassword = "test34test!23"
	seedAmount        = 5 * units.KiloAvax
	stakeAmount       = 3 * units.KiloAvax
	delegatorAmount   = 3 * units.KiloAvax
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
	genesisClient := rpc_workflow_runner.NewRPCWorkFlowRunner(
		e.stakerClient,
		api.UserPass{Username: genesisUsername, Password: genesisPassword},
		e.acceptanceTimeout,
	)

	if _, err := genesisClient.ImportGenesisFunds(); err != nil {
		return stacktrace.Propagate(err, "Failed to fund genesis client.")
	}

	stakerNodeID, err := e.stakerClient.InfoAPI().GetNodeID()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get staker node ID.")
	}
	delegatorNodeID, err := e.delegatorClient.InfoAPI().GetNodeID()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get delegator node ID.")
	}
	highLevelStakerClient := rpc_workflow_runner.NewRPCWorkFlowRunner(
		e.stakerClient,
		api.UserPass{Username: stakerUsername, Password: stakerPassword},
		e.acceptanceTimeout,
	)
	highLevelDelegatorClient := rpc_workflow_runner.NewRPCWorkFlowRunner(
		e.delegatorClient,
		api.UserPass{Username: delegatorUsername, Password: delegatorPassword},
		e.acceptanceTimeout,
	)

	// ====================================== CREATE FUNDED ACCOUNTS ===============================
	stakerXChainAddress, stakerPChainAddress, err := highLevelStakerClient.CreateDefaultAddresses()
	if err != nil {
		return stacktrace.Propagate(err, "Could not create default addresses for staker client.")
	}
	delegatorXChainAddress, delegatorPChainAddress, err := highLevelDelegatorClient.CreateDefaultAddresses()
	if err != nil {
		return stacktrace.Propagate(err, "Could not create default addresses for delegator client.")
	}

	if err := genesisClient.FundXChainAddresses([]string{stakerXChainAddress, delegatorXChainAddress}, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Failed to fund X Chain Addresses from genesis client.")
	}

	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain balance for staker client.")
	}
	if err := highLevelDelegatorClient.VerifyXChainAVABalance(delegatorXChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance for delegator client.")
	}

	//  ====================================== ADD VALIDATOR ===============================
	err = highLevelStakerClient.TransferAvaXChainToPChain(stakerPChainAddress, seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information")
	}
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain balance after X -> P Transfer.")
	}
	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXChainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "X Chain Balance not updated correctly after X -> P Transfer for validator")
	}
	err = highLevelStakerClient.AddValidatorOnSubnet(stakerNodeID, stakerPChainAddress, stakeAmount)
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
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPChainAddress, expectedStakerBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after adding default subnet validator to the network")
	}

	// ====================================== ADD DELEGATOR ======================================
	err = highLevelDelegatorClient.TransferAvaXChainToPChain(delegatorPChainAddress, seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVA from X Chain to P Chain account.")
	}
	if err := highLevelDelegatorClient.VerifyPChainBalance(delegatorPChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain balance after X -> P Transfer for Delegator.")
	}
	if err := highLevelDelegatorClient.VerifyXChainAVABalance(delegatorXChainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance after X -> P Transfer for Delegator")
	}

	err = highLevelDelegatorClient.AddDelegatorOnSubnet(stakerNodeID, delegatorPChainAddress, delegatorAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add delegator %s to default subnet.", delegatorNodeID)
	}
	expectedDelegatorBalance := seedAmount - delegatorAmount
	if err := highLevelDelegatorClient.VerifyPChainBalance(delegatorPChainAddress, expectedDelegatorBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after adding a new delegator to the network.")
	}

	// ====================================== TRANSFER TO X CHAIN ================================
	err = highLevelStakerClient.TransferAvaPChainToXChain(stakerXChainAddress, expectedStakerBalance)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to transfer Ava from P Chain to X Chain.")
	}
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPChainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after P -> X Transfer.")
	}
	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXChainAddress, expectedStakerBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer.")
	}

	err = highLevelDelegatorClient.TransferAvaPChainToXChain(delegatorXChainAddress, expectedStakerBalance)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to transfer Ava from P Chain to X Chain.")
	}
	if err := highLevelDelegatorClient.VerifyPChainBalance(delegatorPChainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after P -> X Transfer.")
	}
	if err := highLevelDelegatorClient.VerifyXChainAVABalance(delegatorXChainAddress, expectedDelegatorBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer.")
	}

	return nil
}
