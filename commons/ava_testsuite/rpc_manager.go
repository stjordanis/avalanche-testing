package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_default_testnet"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	GENESIS_USERNAME = "genesis"
	GENESIS_PASSWORD = "genesis34!23"
	ACCEPTED_STATUS = "Accepted"
)

type RpcManager struct {
	client *gecko_client.GeckoClient
	testNet *ava_default_testnet.TestNet
	rpcUser *RpcUser
}

func NewRpcManager(
		client *gecko_client.GeckoClient,
		testNet *ava_default_testnet.TestNet,
		username string,
		password string) *RpcManager {
	return &RpcManager{
		client: client,
		testNet: testNet,
		rpcUser: NewRpcUser(username, password),
	}
}

type RpcUser struct {
	username string
	password string
	payerNonce int
}

func NewRpcUser(username string, password string) *RpcUser {
	return &RpcUser{username: username, password: password, payerNonce: 0}
}

func (rpcUser RpcUser) incrementNonce() int {
	rpcUser.payerNonce++
	return rpcUser.payerNonce
}

/*
	Creates a new account on the XChain under the username and password.
	Transfers funds from the genesis account to the new XChain account using the Genesis private key.
	Returns the new, funded XChain account address.
 */
func (rpcManager RpcManager) CreateAndSeedXChainAccountFromGenesis(
	username string,
	password string,
	amount int) (string, error) {
	time.Sleep(time.Second * 30)
	client := rpcManager.client
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
		rpcManager.testNet.FundedAddresses.PrivateKey)
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
	txnId, err := client.XChainApi().Send(amount, "AVA", testAccountAddress, GENESIS_USERNAME, GENESIS_PASSWORD)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to send AVA to test account address %s", testAccountAddress)
	}
	err = rpcManager.waitForTransactionAcceptance(txnId)
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
func (rpcManager RpcManager) TransferAvaXChainToPChain(
		username string,
		password string,
		amount int) (string, error) {
	client := rpcManager.client
	pchainAddress, err := client.PChainApi().CreateAccount(username, password, nil)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create new account on PChain")
	}
	txnId, err := client.XChainApi().ExportAVA(pchainAddress, amount, username, password)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed pchainAddress export AVA pchainAddress %s", pchainAddress)
	}
	err = rpcManager.waitForTransactionAcceptance(txnId)
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	txnId, err = client.PChainApi().ImportAVA(username, password, pchainAddress, rpcManager.rpcUser.incrementNonce())
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed pchainAddress import AVA pchainAddress %s", pchainAddress)
	}
	return pchainAddress, nil
}

func (rpcManager RpcManager) waitForTransactionAcceptance(txnId string) error {
	client := rpcManager.client
	status := ""
	for status != ACCEPTED_STATUS {
		status, err := client.XChainApi().GetTxStatus(txnId)
		if err != nil {
			return stacktrace.Propagate(err,"Failed to get status.")
		}
		logrus.Debugf("Status for transaction %s: %s", txnId, status)
		time.Sleep(time.Second)
	}
	return nil
}