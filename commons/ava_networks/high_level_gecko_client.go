package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

const (
	GENESIS_USERNAME            = "genesis"
	GENESIS_PASSWORD            = "genesis34!23"
	TRANSACTION_ACCEPTED_STATUS = "Accepted"
	AVA_ASSET_ID = "AVA"
	TIME_UNTIL_STAKING_BEGINS = 20 * time.Second
	TIME_UNTIL_STAKING_ENDS = 72 * time.Hour
	TIME_UNTIL_DELEGATING_BEGINS = 20 * time.Second
	TIME_UNTIL_DELEGATING_ENDS = 72 * time.Hour
	DELEGATION_FEE_RATE = 500000
	XCHAIN_ADDRESS_PREFIX = "X-"
	NO_IMPORT_INPUTS_ERROR_STR = "problem issuing transaction: no import inputs"
)

type HighLevelGeckoClient struct {
	client    *gecko_client.GeckoClient
	geckoUser *GeckoUser
}

func NewHighLevelGeckoClient(
		client *gecko_client.GeckoClient,
		username string,
		password string) *HighLevelGeckoClient {
	return &HighLevelGeckoClient{
		client:    client,
		geckoUser: NewGeckoUser(username, password),
	}
}

type GeckoUser struct {
	username string
	password string
}

func NewGeckoUser(username string, password string) *GeckoUser {
	return &GeckoUser{username: username, password: password}
}

/*
	High level function that takes a regular node with no Ava and funds it from genesis,
	transfers those funds to the PChain, and registers it as a validator on the default subnet.
 */
func (highLevelGeckoClient HighLevelGeckoClient) GetFundsAndStartValidating(
	    seedAmount int64,
	    stakeAmount int64) error {
	client := highLevelGeckoClient.client
	stakerNodeId, err := client.InfoApi().GetNodeId()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get staker node ID.")
	}
	_, err = highLevelGeckoClient.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not seed XChain account from Genesis.")
	}
	stakerPchainAddress, err := highLevelGeckoClient.TransferAvaXChainToPChain(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVA from XChain to PChain account information")
	}
	_, err = highLevelGeckoClient.CreateAndSeedXChainAccountFromGenesis(seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not seed XChain account from Genesis.")
	}
	// Adding staker
	err = highLevelGeckoClient.AddValidatorOnSubnet(stakerNodeId, stakerPchainAddress, stakeAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add staker %s to default subnet.", stakerNodeId)
	}
	return nil
}

func (highLevelGeckoClient HighLevelGeckoClient) AddDelegatorOnSubnet(
		delegateeNodeId string,
		pchainAddress string,
		stakeAmount int64,
		) error {
	client := highLevelGeckoClient.client
	currentPayerNonce, err := highLevelGeckoClient.getCurrentPayerNonce(pchainAddress)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get payer nonce from address %s", pchainAddress)
	}
	delegatorStartTime := time.Now().Add(TIME_UNTIL_DELEGATING_BEGINS).Unix()
	addDelegatorUnsignedTxn, err := client.PChainApi().AddDefaultSubnetDelegator(
		delegateeNodeId,
		delegatorStartTime,
		time.Now().Add(TIME_UNTIL_DELEGATING_ENDS).Unix(),
		stakeAmount,
		currentPayerNonce + 1,
		pchainAddress)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to add default subnet delegator %s", pchainAddress)
	}
	addDelegatorSignedTxn, err := client.PChainApi().Sign(
		addDelegatorUnsignedTxn,
		pchainAddress,
		highLevelGeckoClient.geckoUser.username,
		highLevelGeckoClient.geckoUser.password)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to sign delegator transaction.")
	}
	_, err = client.PChainApi().IssueTx(addDelegatorSignedTxn)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue staker transaction.")
	}
	for time.Now().Unix() < delegatorStartTime {
		time.Sleep(time.Second)
	}
	return nil
}

func (highLevelGeckoClient HighLevelGeckoClient) AddValidatorOnSubnet(
		nodeId string,
		pchainAddress string,
		stakeAmount int64) error {
	client := highLevelGeckoClient.client
	currentPayerNonce, err := highLevelGeckoClient.getCurrentPayerNonce(pchainAddress)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get payer nonce from address %s", pchainAddress)
	}
	stakingStartTime := time.Now().Add(TIME_UNTIL_STAKING_BEGINS).Unix()
	addStakerUnsignedTxn, err := client.PChainApi().AddDefaultSubnetValidator(
		nodeId,
		stakingStartTime,
		time.Now().Add(TIME_UNTIL_STAKING_ENDS).Unix(),
		stakeAmount,
		currentPayerNonce + 1,
		pchainAddress,
		DELEGATION_FEE_RATE)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to add default subnet staker %s", nodeId)
	}
	addStakerSignedTxn, err := client.PChainApi().Sign(
		addStakerUnsignedTxn,
		pchainAddress,
		highLevelGeckoClient.geckoUser.username,
		highLevelGeckoClient.geckoUser.password)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to sign staker transaction.")
	}
	_, err = client.PChainApi().IssueTx(addStakerSignedTxn)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue staker transaction.")
	}
	for time.Now().Unix() < stakingStartTime {
		time.Sleep(time.Second)
	}
	return nil
}


/*
	Creates a new account on the XChain under the username and password.
	Transfers funds from the genesis account to the new XChain account using the Genesis private key.
	Returns the new, funded XChain account address.
 */
func (highLevelGeckoClient HighLevelGeckoClient) CreateAndSeedXChainAccountFromGenesis(
		amount int64) (string, error) {
	client := highLevelGeckoClient.client
	username := highLevelGeckoClient.geckoUser.username
	password := highLevelGeckoClient.geckoUser.password
	_, err := client.KeystoreApi().CreateUser(username, password)
	if err != nil {
		stacktrace.Propagate(err, "Could not create user.")
	}
	_, err = client.KeystoreApi().CreateUser(GENESIS_USERNAME, GENESIS_PASSWORD)
	if err != nil {
		stacktrace.Propagate(err, "Could not create genesis user.")
	}
	nodeId, err := client.InfoApi().GetNodeId()
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not get node id")
	}
	genesisAccountAddress, err := client.XChainApi().ImportKey(
		GENESIS_USERNAME,
		GENESIS_PASSWORD,
		DefaultLocalNetGenesisConfig.FundedAddresses.PrivateKey)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to take control of genesis account.")
	}
	logrus.Debugf("Adding Node %s as a validator.", nodeId)
	logrus.Debugf("Genesis Address: %s.", genesisAccountAddress)
	testAccountAddress, err := client.XChainApi().CreateAddress(username, password)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create address on XChain.")
	}
	logrus.Debugf("Test account address: %s", testAccountAddress)
	txnId, err := client.XChainApi().Send(amount, AVA_ASSET_ID, testAccountAddress, GENESIS_USERNAME, GENESIS_PASSWORD)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to send AVA to test account address %s", testAccountAddress)
	}
	err = highLevelGeckoClient.waitForXchainTransactionAcceptance(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to wait for transaction acceptance.")
	}
	return testAccountAddress, nil
}

/*
	Creates a new account on the PChain under the username and password.
	Transfers funds from an XChain account owned by that username and password to the new PChain account.
	Returns the new, funded PChain account address.
*/
func (highLevelGeckoClient HighLevelGeckoClient) TransferAvaXChainToPChain(
		amount int64) (string, error) {
	client := highLevelGeckoClient.client
	username := highLevelGeckoClient.geckoUser.username
	password := highLevelGeckoClient.geckoUser.password
	pchainAddress, err := client.PChainApi().CreateAccount(username, password, nil)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create new account on PChain")
	}
	txnId, err := client.XChainApi().ExportAVA(pchainAddress, amount, username, password)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to export AVA to pchainAddress %s", pchainAddress)
	}
	err = highLevelGeckoClient.waitForXchainTransactionAcceptance(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	currentPayerNonce, err := highLevelGeckoClient.getCurrentPayerNonce(pchainAddress)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to get payer nonce from address %s", pchainAddress)
	}
	txnId, err = client.PChainApi().ImportAVA(username, password, pchainAddress, currentPayerNonce + 1)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed import AVA to pchainAddress %s", pchainAddress)
	}
	txnId, err = client.PChainApi().IssueTx(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to issue importAVA transaction.")
	}
	highLevelGeckoClient.waitForPchainNonZeroBalance(pchainAddress)
	return pchainAddress, nil
}

/*
	Transfers funds from a Phain account owned by that username and password to an XChain account.
	Returns the XChain account address.
*/
func (highLevelGeckoClient HighLevelGeckoClient) TransferAvaPChainToXChain(
	// HighLevelGeckoClient must own both pchainAddress and xchainAddress.
		pchainAddress string,
		xchainAddress string,
		amount int64) (string, error) {
	client := highLevelGeckoClient.client
	username := highLevelGeckoClient.geckoUser.username
	password := highLevelGeckoClient.geckoUser.password
	xchainAddressWithoutPrefix := strings.TrimPrefix(xchainAddress, XCHAIN_ADDRESS_PREFIX)
	currentPayerNonce, err := highLevelGeckoClient.getCurrentPayerNonce(pchainAddress)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to get current payer nonce from pchainAddress %v", pchainAddress)
	}
	// PChain API only accepts the XChain address without the xchain prefix.
	unsignedTxnId, err := client.PChainApi().ExportAVA(amount, xchainAddressWithoutPrefix, currentPayerNonce + 1)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to export AVA to xchainAddress %s", xchainAddress)
	}
	signedTxnId, err := client.PChainApi().Sign(unsignedTxnId, pchainAddress, username, password)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to sign export AVA transaction.")
	}
	_, err = client.PChainApi().IssueTx(signedTxnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to issue importAVA transaction.")
	}
	// XChain API only accepts the XChain address with the xchain prefix.
	txnId, err := client.XChainApi().ImportAVA(xchainAddress, username, password)
	for err != nil {
		/*
			HACK HACK HACK because the PChain does not have a way to verify transaction acceptence yet,
			we retry based on the contents of the error message from the XChain call if the pchain transaction
			has not yet reached consensus
		*/
		// TODO TODO TODO When the PChain transaction status endpoint is deployed, use that to wait fro transaction acceptance
		if strings.Contains(err.Error(), NO_IMPORT_INPUTS_ERROR_STR) {
			txnId, err = client.XChainApi().ImportAVA(xchainAddress, username, password)
			time.Sleep(time.Second)
		} else {
			return "", stacktrace.Propagate(err, "Failed import AVA to xchainAddress %s", xchainAddress)
		}
	}
	err = highLevelGeckoClient.waitForXchainTransactionAcceptance(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to wait for acceptance of transaction on XChain.")
	}
	return xchainAddress, nil
}

func (highLevelGeckoClient HighLevelGeckoClient) waitForXchainTransactionAcceptance(txnId string) error {
	client := highLevelGeckoClient.client
	status, err := client.XChainApi().GetTxStatus(txnId)
	if err != nil {
		return stacktrace.Propagate(err,"Failed to get status.")
	}
	for status != TRANSACTION_ACCEPTED_STATUS {
		status, err = client.XChainApi().GetTxStatus(txnId)
		if err != nil {
			return stacktrace.Propagate(err,"Failed to get status.")
		}
		logrus.Debugf("Status for transaction %s: %s", txnId, status)
		time.Sleep(time.Second)
	}
	return nil
}

func (highLevelGeckoClient HighLevelGeckoClient) waitForPchainNonZeroBalance(pchainAddress string) error {
	client := highLevelGeckoClient.client
	pchainAccount, err := client.PChainApi().GetAccount(pchainAddress)
	if err != nil {
		return stacktrace.Propagate(err, "Could not get PChain account information")
	}
	balance := pchainAccount.Balance
	if err != nil {
		return stacktrace.Propagate(err,"Failed to get balance.")
	}
	for balance == "0" {
		pchainAccount, err = client.PChainApi().GetAccount(pchainAddress)
		if err != nil {
			return stacktrace.Propagate(err,"Failed to get account information.")
		}
		balance = pchainAccount.Balance
		logrus.Debugf("Balance for account %s: %s", pchainAddress, balance)
		time.Sleep(time.Second)
	}
	return nil
}

func (highLevelGeckoClient HighLevelGeckoClient) getCurrentPayerNonce(pchainAddress string) (int, error) {
	pchainAccountInfo, err := highLevelGeckoClient.client.PChainApi().GetAccount(pchainAddress)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to get pchain account info.")
	}
	currentPayerNonce, err := strconv.Atoi(pchainAccountInfo.Nonce)
	if err != nil {
		return 0, stacktrace.Propagate(err, "")
	}
	return currentPayerNonce, nil
}