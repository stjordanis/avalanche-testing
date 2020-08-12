package rpc_workflow_runner

import (
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_networks"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/ava-labs/gecko/vms/platformvm"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	AVA_ASSET_ID                = "AVAX"
	DefaultStakingDelay         = 20 * time.Second
	DefaultStakingPeriod        = 72 * time.Hour
	DefaultDelegationDelay      = 20 * time.Second // Time until delegation period should begin
	stakingPeriodSynchronyDelay = 3 * time.Second
	DefaultDelegationPeriod     = 36 * time.Hour
	DefaultDelegationFeeRate    = 500000
)

/*
	RpcWorkflowRunner executes standard testing workflows like funding accounts from
	genesis and adding nodes as validators, using the a given gecko client handle as the
	entry point to the test network. It runs the RpcWorkflows using the credential
	set in the GeckoUser field.
*/
type RpcWorkflowRunner struct {
	client    *apis.Client
	geckoUser api.UserPass
	/*
		This timeout represents the time the RpcWorkflowRunner will wait for some state
		change in the network to be understood as accepted and implemented by the underlying
		Gecko client (XChain transaction acceptance, Ava transfer to PChain, etc). There is
		only one timeout for each kind of state change in order to reduce the complexity of
		configuring timeouts throughout the test suite.
		Also, each state change is roughly the same - we're waiting not only for
		a transaction to be considered accepted by the network and also for the nodes
		internal state to reflect that acceptance.
	*/
	networkAcceptanceTimeout time.Duration
}

func NewRpcWorkflowRunner(
	client *apis.Client,
	username string,
	password string,
	networkAcceptanceTimeout time.Duration) *RpcWorkflowRunner {
	return &RpcWorkflowRunner{
		client: client,
		geckoUser: api.UserPass{
			Username: username,
			Password: password,
		},
		networkAcceptanceTimeout: networkAcceptanceTimeout,
	}
}

/*
	High level function that takes a regular node with no Ava and funds it from genesis,
	transfers those funds to the PChain, and registers it as a validator on the default subnet.
*/
func (runner RpcWorkflowRunner) GetFundsAndStartValidating(
	seedAmount uint64,
	stakeAmount uint64) error {
	client := runner.client
	stakerNodeId, err := client.InfoAPI().GetNodeID()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get staker node ID.")
	}
	_, err = runner.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not seed XChain account from Genesis.")
	}
	stakerPchainAddress, err := runner.TransferAvaXChainToPChain(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information")
	}
	_, err = runner.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not seed XChain account from Genesis.")
	}
	// Adding staker
	err = runner.AddValidatorOnSubnet(stakerNodeId, stakerPchainAddress, stakeAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add staker %s to default subnet.", stakerNodeId)
	}
	return nil
}

func (runner RpcWorkflowRunner) AddDelegatorOnSubnet(
	delegateeNodeId string,
	pchainAddress string,
	stakeAmount uint64,
) error {
	client := runner.client
	delegatorStartTime := time.Now().Add(DefaultDelegationDelay)
	startTime := uint64(delegatorStartTime.Unix())
	endTime := uint64(delegatorStartTime.Add(DefaultDelegationPeriod).Unix())
	addDelegatorTxID, err := client.PChainAPI().AddDefaultSubnetDelegator(
		runner.geckoUser,
		pchainAddress,
		delegateeNodeId,
		stakeAmount,
		startTime,
		endTime,
	)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to add default subnet delegator %s", pchainAddress)
	}
	if err := runner.waitForPChainTransactionAcceptance(addDelegatorTxID); err != nil {
		return stacktrace.Propagate(err, "Failed to accept AddDefaultSubnetDelegator tx: %s", addDelegatorTxID)
	}

	// Sleep until delegator starts validating
	time.Sleep(time.Until(delegatorStartTime) + stakingPeriodSynchronyDelay)
	return nil
}

func (runner RpcWorkflowRunner) AddValidatorOnSubnet(
	nodeId string,
	pchainAddress string,
	stakeAmount uint64) error {
	// Replace with simple call to AddDefaultSubnetValidator
	client := runner.client
	stakingStartTime := time.Now().Add(DefaultStakingDelay)
	startTime := uint64(stakingStartTime.Unix())
	endTime := uint64(stakingStartTime.Add(DefaultStakingPeriod).Unix())
	addStakerTxID, err := client.PChainAPI().AddDefaultSubnetValidator(
		runner.geckoUser,
		pchainAddress,
		nodeId,
		stakeAmount,
		startTime,
		endTime,
		DefaultDelegationFeeRate,
	)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to add default subnet staker %s", nodeId)
	}

	if err := runner.waitForPChainTransactionAcceptance(addStakerTxID); err != nil {
		return stacktrace.Propagate(err, "Failed to confirm AddDefaultSubnetValidator Tx: %s", addStakerTxID)
	}

	time.Sleep(time.Until(stakingStartTime) + stakingPeriodSynchronyDelay)

	return nil
}

/*
	Creates a new account on the XChain under the username and password.
	Transfers funds from the genesis account to the new XChain account using the Genesis private key.
	Returns the new, funded XChain account address.
*/
func (runner RpcWorkflowRunner) CreateAndSeedXChainAccountFromGenesis(
	amount uint64) (string, error) {
	client := runner.client
	keystore := client.KeystoreAPI()
	_, err := keystore.CreateUser(runner.geckoUser)
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not create user.")
	}
	nodeId, err := client.InfoAPI().GetNodeID()
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not get node id")
	}
	genesisAccountAddress, err := client.XChainAPI().ImportKey(
		runner.geckoUser,
		ava_networks.DefaultLocalNetGenesisConfig.FundedAddresses.PrivateKey)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to take control of genesis account.")
	}
	logrus.Debugf("Adding Node %s as a validator.", nodeId)
	logrus.Debugf("Genesis Address: %s.", genesisAccountAddress)
	testAccountAddress, err := client.XChainAPI().CreateAddress(runner.geckoUser)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create address on XChain.")
	}
	logrus.Debugf("Test account address: %s", testAccountAddress)

	txnId, err := client.XChainAPI().Send(runner.geckoUser, amount, AVA_ASSET_ID, testAccountAddress)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to send AVA to test account address %s", testAccountAddress)
	}
	err = runner.waitForXchainTransactionAcceptance(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to wait for transaction acceptance.")
	}
	return testAccountAddress, nil
}

/*
	Creates a new account on the PChain for geckoUser
	Transfers funds from an XChain account owned by geckoUser to the new PChain account.
	Returns the new, funded PChain account address.
*/
func (runner RpcWorkflowRunner) TransferAvaXChainToPChain(
	amount uint64) (string, error) {
	client := runner.client
	pchainAddress, err := client.PChainAPI().CreateAddress(runner.geckoUser)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create new address on PChain")
	}

	txnId, err := client.XChainAPI().ExportAVAX(runner.geckoUser, amount, pchainAddress)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to export AVA to pchainAddress %s", pchainAddress)
	}
	err = runner.waitForXchainTransactionAcceptance(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}

	importTxID, err := client.PChainAPI().ImportAVAX(runner.geckoUser, pchainAddress)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed import AVA to pchainAddress %s", pchainAddress)
	}
	if err := runner.waitForPChainTransactionAcceptance(importTxID); err != nil {
		return "", stacktrace.Propagate(err, "Failed to Accept ImportTx: %s", importTxID)
	}

	return pchainAddress, nil
}

/*
	Transfers funds from a PChain account owned by geckoUser to an XChain account.
	Returns the XChain account address.
*/
func (runner RpcWorkflowRunner) TransferAvaPChainToXChain(
	// RpcWorkflowRunner must own both pchainAddress and xchainAddress.
	pchainAddress string,
	xchainAddress string,
	amount uint64) (string, error) {
	client := runner.client

	exportTxID, err := client.PChainAPI().ExportAVAX(runner.geckoUser, xchainAddress, amount)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to export AVA to xchainAddress %s", xchainAddress)
	}
	if err := runner.waitForPChainTransactionAcceptance(exportTxID); err != nil {
		return "", stacktrace.Propagate(err, "Failed to accept ExportTx: %s", exportTxID)
	}

	txnId, err := client.XChainAPI().ImportAVAX(runner.geckoUser, xchainAddress)
	err = runner.waitForXchainTransactionAcceptance(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to wait for acceptance of transaction on XChain.")
	}
	return xchainAddress, nil
}

func (runner RpcWorkflowRunner) waitForXchainTransactionAcceptance(txnId ids.ID) error {
	client := runner.client
	status, err := client.XChainAPI().GetTxStatus(txnId)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get status.")
	}
	pollStartTime := time.Now()
	for time.Since(pollStartTime) < runner.networkAcceptanceTimeout && status != choices.Accepted {
		status, err = client.XChainAPI().GetTxStatus(txnId)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get status.")
		}
		logrus.Debugf("Status for transaction %s: %s", txnId, status)
		time.Sleep(time.Second)
	}
	if status != choices.Accepted {
		return stacktrace.NewError("Timed out waiting for transaction %s to be accepted on the XChain.", txnId)
	} else {
		return nil
	}
}

func (runner RpcWorkflowRunner) waitForPChainTransactionAcceptance(txID ids.ID) error {
	client := runner.client.PChainAPI()
	pollStartTime := time.Now()

	status, err := client.GetTxStatus(txID)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get tx status")
	}

	for time.Since(pollStartTime) < runner.networkAcceptanceTimeout && status != platformvm.Committed {
		time.Sleep(2 * time.Second)
		status, err = client.GetTxStatus(txID)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get status")
		}
		logrus.Debugf("Status for transaction: %s: %s", txID, status)
		if status == platformvm.Dropped {
			return stacktrace.NewError("Transaction %s was dropped", txID)
		}
	}
	if status != platformvm.Committed {
		return stacktrace.NewError("Timed out waiting for transaction %s to be accepted on the PChain.", txID)
	}

	return nil
}
