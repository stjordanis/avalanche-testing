package ava_networks

import (
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	GENESIS_USERNAME            = "genesis"
	GENESIS_PASSWORD            = "genesis34!23"
	TRANSACTION_ACCEPTED_STATUS = "Accepted"
	AVA_ASSET_ID = "AVA"
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
	nodeId, err := client.AdminApi().GetNodeId()
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
	err = highLevelGeckoClient.waitForTransactionAcceptance(txnId)
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
	err = highLevelGeckoClient.waitForTransactionAcceptance(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	pchainAccountInfo, err := client.PChainApi().GetAccount(pchainAddress)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to get pchain account info.")
	}
	currentPayerNonce, err := strconv.Atoi(pchainAccountInfo.Nonce)
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	txnId, err = client.PChainApi().ImportAVA(username, password, pchainAddress, currentPayerNonce + 1)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed import AVA to pchainAddress %s", pchainAddress)
	}
	txnId, err = client.PChainApi().IssueTx(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to issue importAVA transaction.")
	}
	highLevelGeckoClient.waitForNonZeroBalance(pchainAddress)
	return pchainAddress, nil
}

func (highLevelGeckoClient HighLevelGeckoClient) waitForTransactionAcceptance(txnId string) error {
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

func (highLevelGeckoClient HighLevelGeckoClient) waitForNonZeroBalance(pchainAddress string) error {
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