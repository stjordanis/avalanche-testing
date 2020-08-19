package rpc_workflow_runner

import (
	"time"

	avalancheNetwork "github.com/ava-labs/avalanche-e2e-tests/commons/ava_networks"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/utils/constants"
	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/ava-labs/gecko/vms/platformvm"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	AvaxAssetID                 = "AVAX"
	DefaultStakingDelay         = 20 * time.Second
	DefaultStakingPeriod        = 72 * time.Hour
	DefaultDelegationDelay      = 20 * time.Second // Time until delegation period should begin
	stakingPeriodSynchronyDelay = 3 * time.Second
	DefaultDelegationPeriod     = 36 * time.Hour
	DefaultDelegationFeeRate    = 0.1
)

// RPCWorkFlowRunner executes standard testing workflows like funding accounts from
// genesis and adding nodes as validators, using the a given gecko client handle as the
// entry point to the test network. It runs the RpcWorkflows using the credential
// set in the GeckoUser field.
type RPCWorkFlowRunner struct {
	client    *apis.Client
	geckoUser api.UserPass

	// This timeout represents the time the RPCWorkFlowRunner will wait for some state change to be accepted
	// and implemented by the underlying client.
	networkAcceptanceTimeout time.Duration
}

// NewRPCWorkFlowRunner ...
func NewRPCWorkFlowRunner(
	client *apis.Client,
	user api.UserPass,
	networkAcceptanceTimeout time.Duration) *RPCWorkFlowRunner {
	return &RPCWorkFlowRunner{
		client:                   client,
		geckoUser:                user,
		networkAcceptanceTimeout: networkAcceptanceTimeout,
	}
}

// ImportGenesisFunds imports the genesis private key to this user's keystore
func (runner RPCWorkFlowRunner) ImportGenesisFunds() (string, error) {
	client := runner.client
	keystore := client.KeystoreAPI()
	if _, err := keystore.CreateUser(runner.geckoUser); err != nil {
		return "", err
	}

	genesisAccountAddress, err := client.XChainAPI().ImportKey(
		runner.geckoUser,
		avalancheNetwork.DefaultLocalNetGenesisConfig.FundedAddresses.PrivateKey)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to take control of genesis account.")
	}
	logrus.Debugf("Genesis Address: %s.", genesisAccountAddress)
	return genesisAccountAddress, nil
}

// ImportGenesisFundsAndStartValidating attempts to import genesis funds and add this node as a validator
func (runner RPCWorkFlowRunner) ImportGenesisFundsAndStartValidating(
	seedAmount uint64,
	stakeAmount uint64) (string, error) {
	client := runner.client
	stakerNodeID, err := client.InfoAPI().GetNodeID()
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not get staker node ID.")
	}
	_, err = runner.ImportGenesisFunds()
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not seed XChain account from Genesis.")
	}
	pChainAddress, err := client.PChainAPI().CreateAddress(runner.geckoUser)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create new address on PChain")
	}
	err = runner.TransferAvaXChainToPChain(pChainAddress, seedAmount)
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information")
	}
	// Adding staker
	err = runner.AddValidatorOnSubnet(stakerNodeID, pChainAddress, stakeAmount)
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not add staker %s to default subnet.", stakerNodeID)
	}
	return pChainAddress, nil
}

// AddDelegatorOnSubnet delegates to [delegateeNodeID] and blocks until the transaction is confirmed and the delegation
// period begins
func (runner RPCWorkFlowRunner) AddDelegatorOnSubnet(
	delegateeNodeID string,
	pChainAddress string,
	stakeAmount uint64,
) error {
	client := runner.client
	delegatorStartTime := time.Now().Add(DefaultDelegationDelay)
	startTime := uint64(delegatorStartTime.Unix())
	endTime := uint64(delegatorStartTime.Add(DefaultDelegationPeriod).Unix())
	addDelegatorTxID, err := client.PChainAPI().AddDefaultSubnetDelegator(
		runner.geckoUser,
		pChainAddress,
		delegateeNodeID,
		stakeAmount,
		startTime,
		endTime,
	)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to add default subnet delegator %s", pChainAddress)
	}
	if err := runner.waitForPChainTransactionAcceptance(addDelegatorTxID); err != nil {
		return stacktrace.Propagate(err, "Failed to accept AddDefaultSubnetDelegator tx: %s", addDelegatorTxID)
	}

	// Sleep until delegator starts validating
	time.Sleep(time.Until(delegatorStartTime) + stakingPeriodSynchronyDelay)
	return nil
}

// AddValidatorOnSubnet adds [nodeID] as a validator and blocks until the transaction is confirmed and the validation
// period begins
func (runner RPCWorkFlowRunner) AddValidatorOnSubnet(
	nodeID string,
	pchainAddress string,
	stakeAmount uint64,
) error {
	// Replace with simple call to AddDefaultSubnetValidator
	client := runner.client
	stakingStartTime := time.Now().Add(DefaultStakingDelay)
	startTime := uint64(stakingStartTime.Unix())
	endTime := uint64(stakingStartTime.Add(DefaultStakingPeriod).Unix())
	addStakerTxID, err := client.PChainAPI().AddDefaultSubnetValidator(
		runner.geckoUser,
		pchainAddress,
		nodeID,
		stakeAmount,
		startTime,
		endTime,
		DefaultDelegationFeeRate,
	)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to add default subnet staker %s", nodeID)
	}

	if err := runner.waitForPChainTransactionAcceptance(addStakerTxID); err != nil {
		return stacktrace.Propagate(err, "Failed to confirm AddDefaultSubnetValidator Tx: %s", addStakerTxID)
	}

	time.Sleep(time.Until(stakingStartTime) + stakingPeriodSynchronyDelay)

	return nil
}

// FundXChainAddresses sends [amount] AVA to each address in [addresses] and returns the created txIDs
func (runner RPCWorkFlowRunner) FundXChainAddresses(addresses []string, amount uint64) error {
	client := runner.client.XChainAPI()
	for _, address := range addresses {
		txID, err := client.Send(runner.geckoUser, amount, AvaxAssetID, address)
		if err != nil {
			return err
		}
		if err := runner.waitForXchainTransactionAcceptance(txID); err != nil {
			return err
		}
	}

	return nil
}

// CreateDefaultAddresses creates the keystore user for this workflow runner and
// creates an X and P Chain address for that keystore user
func (runner RPCWorkFlowRunner) CreateDefaultAddresses() (string, string, error) {
	client := runner.client
	keystore := client.KeystoreAPI()
	if _, err := keystore.CreateUser(runner.geckoUser); err != nil {
		return "", "", err
	}

	xAddress, err := client.XChainAPI().CreateAddress(runner.geckoUser)
	if err != nil {
		return "", "", err
	}

	pAddress, err := client.PChainAPI().CreateAddress(runner.geckoUser)
	return xAddress, pAddress, err
}

// TransferAvaXChainToPChain exports AVA from the X Chain and then imports it to the P Chain
// and blocks until both transactions have been accepted
func (runner RPCWorkFlowRunner) TransferAvaXChainToPChain(pChainAddress string, amount uint64) error {
	client := runner.client
	txID, err := client.XChainAPI().ExportAVAX(runner.geckoUser, amount, pChainAddress)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to export AVA to pchainAddress %s", pChainAddress)
	}
	err = runner.waitForXchainTransactionAcceptance(txID)
	if err != nil {
		return stacktrace.Propagate(err, "")
	}

	importTxID, err := client.PChainAPI().ImportAVAX(runner.geckoUser, pChainAddress, constants.XChainID.String())
	if err != nil {
		return stacktrace.Propagate(err, "Failed import AVA to pchainAddress %s", pChainAddress)
	}
	if err := runner.waitForPChainTransactionAcceptance(importTxID); err != nil {
		return stacktrace.Propagate(err, "Failed to Accept ImportTx: %s", importTxID)
	}

	return nil
}

// TransferAvaPChainToXChain exports AVA from the P Chain and then imports it to the X Chain
// and blocks until both transactions have been accepted
func (runner RPCWorkFlowRunner) TransferAvaPChainToXChain(
	xChainAddress string,
	amount uint64) error {
	client := runner.client

	exportTxID, err := client.PChainAPI().ExportAVAX(runner.geckoUser, xChainAddress, amount)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to export AVA to xChainAddress %s", xChainAddress)
	}
	if err := runner.waitForPChainTransactionAcceptance(exportTxID); err != nil {
		return stacktrace.Propagate(err, "Failed to accept ExportTx: %s", exportTxID)
	}

	txID, err := client.XChainAPI().ImportAVAX(runner.geckoUser, xChainAddress, constants.PlatformChainID.String())
	err = runner.waitForXchainTransactionAcceptance(txID)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to wait for acceptance of transaction on XChain.")
	}
	return nil
}

// waitForXChainTransactionAcceptance gets the status of [txID] and keeps querying until it
// has been accepted
func (runner RPCWorkFlowRunner) waitForXchainTransactionAcceptance(txID ids.ID) error {
	client := runner.client.XChainAPI()

	pollStartTime := time.Now()
	for time.Since(pollStartTime) < runner.networkAcceptanceTimeout {
		status, err := client.GetTxStatus(txID)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get status.")
		}
		if status == choices.Accepted {
			return nil
		}
		if status == choices.Rejected {
			return stacktrace.NewError("Transaciton %s was rejected", txID)
		}
		logrus.Debugf("Status for transaction %s: %s", txID, status)
		time.Sleep(time.Second)
	}

	return stacktrace.NewError("Timed out waiting for transaction %s to be accepted on the XChain.", txID)
}

// waitForPChainTransactionAcceptance gets the status of [txID] and keeps querying until it
// has been accepted
func (runner RPCWorkFlowRunner) waitForPChainTransactionAcceptance(txID ids.ID) error {
	client := runner.client.PChainAPI()
	pollStartTime := time.Now()

	for time.Since(pollStartTime) < runner.networkAcceptanceTimeout {
		status, err := client.GetTxStatus(txID)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get status")
		}
		if status == platformvm.Committed {
			return nil
		}

		// TODO reset to debug log level
		logrus.Infof("Status for transaction: %s: %s", txID, status)
		if status == platformvm.Dropped || status == platformvm.Aborted {
			return stacktrace.NewError("Abandoned Tx: %s because it had status: %s", txID, status)
		}
		time.Sleep(time.Second)
	}

	return stacktrace.NewError("Timed out waiting for transaction %s to be accepted on the PChain.", txID)
}

// VerifyPChainBalance verifies that the balance of P Chain Address: [address] is [expectedBalance]
func (runner RPCWorkFlowRunner) VerifyPChainBalance(address string, expectedBalance uint64) error {
	client := runner.client.PChainAPI()
	balance, err := client.GetBalance(address)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to retrieve P Chain balance.")
	}
	actualBalance := uint64(balance.Balance)
	if actualBalance != expectedBalance {
		return stacktrace.NewError("Found unexpected P Chain Balance for address: %s. Expected: %v, found: %v", address, expectedBalance, actualBalance)
	}

	return nil
}

// VerifyXChainAVABalance verifies that the balance of X Chain Address: [address] is [expectedBalance]
func (runner RPCWorkFlowRunner) VerifyXChainAVABalance(address string, expectedBalance uint64) error {
	client := runner.client.XChainAPI()
	balance, err := client.GetBalance(address, AvaxAssetID)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to retrieve X Chain balance.")
	}
	actualBalance := uint64(balance.Balance)
	if actualBalance != expectedBalance {
		return stacktrace.NewError("Found unexpected X Chain Balance for address: %s. Expected: %v, found: %v", address, expectedBalance, actualBalance)
	}

	return nil
}
