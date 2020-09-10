package workflow

import (
	"time"

	"github.com/ava-labs/avalanche-testing/gecko_client/apis"
	"github.com/ava-labs/avalanche-testing/testsuite/helpers"
	"github.com/ava-labs/avalanche-testing/testsuite/tester"
	"github.com/ava-labs/avalanche-go/api"
	"github.com/ava-labs/avalanche-go/utils/constants"
	"github.com/ava-labs/avalanche-go/utils/units"
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
func NewRPCWorkflowTestExecutor(stakerClient, delegatorClient *apis.Client, acceptanceTimeout time.Duration) tester.AvalancheTester {
	return &executor{
		stakerClient:      stakerClient,
		delegatorClient:   delegatorClient,
		acceptanceTimeout: acceptanceTimeout,
	}
}

// ExecuteTest ...
func (e *executor) ExecuteTest() error {
	genesisClient := helpers.NewRPCWorkFlowRunner(
		e.stakerClient,
		api.UserPass{Username: genesisUsername, Password: genesisPassword},
		e.acceptanceTimeout,
	)

	if _, err := genesisClient.ImportGenesisFunds(); err != nil {
		return stacktrace.Propagate(err, "Failed to fund genesis client.")
	}
	logrus.Debugf("Funded genesis client...")

	stakerNodeID, err := e.stakerClient.InfoAPI().GetNodeID()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get staker node ID.")
	}
	delegatorNodeID, err := e.delegatorClient.InfoAPI().GetNodeID()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get delegator node ID.")
	}
	highLevelStakerClient := helpers.NewRPCWorkFlowRunner(
		e.stakerClient,
		api.UserPass{Username: stakerUsername, Password: stakerPassword},
		e.acceptanceTimeout,
	)
	highLevelDelegatorClient := helpers.NewRPCWorkFlowRunner(
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
	logrus.Infof("Created addresses for staker and delegator clients.")

	if err := genesisClient.FundXChainAddresses([]string{stakerXChainAddress, delegatorXChainAddress}, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Failed to fund X Chain Addresses from genesis client.")
	}

	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain balance for staker client.")
	}
	if err := highLevelDelegatorClient.VerifyXChainAVABalance(delegatorXChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance for delegator client.")
	}
	logrus.Infof("Funded X Chain Addresses for staker and delegator clients.")

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
	err = highLevelStakerClient.AddValidatorToPrimaryNetwork(stakerNodeID, stakerPChainAddress, stakeAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add staker %s to primary network.", stakerNodeID)
	}
	logrus.Infof("Transferred funds from X Chain to P Chain and added a new staker.")

	// ====================================== VERIFY NETWORK STATE ===============================
	currentStakers, currentDelegators, err := e.stakerClient.PChainAPI().GetCurrentValidators(constants.PrimaryNetworkID)
	if err != nil {
		return stacktrace.Propagate(err, "Could not get current stakers.")
	}
	actualNumStakers := len(currentStakers)
	logrus.Debugf("Number of current stakers: %d", actualNumStakers)
	expectedNumStakers := 6
	if actualNumStakers != expectedNumStakers {
		return stacktrace.NewError("Actual number of stakers, %v, != expected number of stakers, %v", actualNumStakers, expectedNumStakers)
	}
	actualNumDelegators := len(currentDelegators)
	logrus.Debugf("Number of current delegators: %d", actualNumDelegators)
	expectedNumDelegators := 0
	if actualNumDelegators != expectedNumDelegators {
		return stacktrace.NewError("Actual number of delegators, %v, != expected number of delegators, %v", actualNumDelegators, expectedNumDelegators)
	}
	expectedStakerBalance := seedAmount - stakeAmount
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPChainAddress, expectedStakerBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after adding  validator to the primary network")
	}
	logrus.Infof("Verified the staker was added to current validators and has the expected P Chain balance.")

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

	err = highLevelDelegatorClient.AddDelegatorToPrimaryNetwork(stakerNodeID, delegatorPChainAddress, delegatorAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add delegator %s to the primary network.", delegatorNodeID)
	}
	expectedDelegatorBalance := seedAmount - delegatorAmount
	if err := highLevelDelegatorClient.VerifyPChainBalance(delegatorPChainAddress, expectedDelegatorBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after adding a new delegator to the network.")
	}
	logrus.Infof("Added delegator to subnet and verified the expected P Chain balance.")

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
	logrus.Infof("Transferred leftover staker funds back to X Chain and verified X and P balances.")

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
	logrus.Infof("Transferred leftover delegator funds back to X Chain and verified X and P balances.")

	return nil
}
