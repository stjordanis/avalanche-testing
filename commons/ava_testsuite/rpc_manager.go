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
	testNet ava_default_testnet.TestNet
}

/*
	Creates a new account on the XChain under the username and password.
	Transfers funds from the genesis account to the new XChain account using the Genesis private key.
	Returns the new, funded XChain account address.
 */
func (rpcManager RpcManager) createAndSeedXChainAccountFromGenesis(
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
	status := ""
	for status != ACCEPTED_STATUS {
		status, err = client.XChainApi().GetTxStatus(txnId)
		if err != nil {
			return "", stacktrace.Propagate(err,"Failed to get status.")
		}
		time.Sleep(time.Second)
	}
	logrus.Debugf("Transaction status for send transaction: %s", status)
	return testAccountAddress, nil
}

/*
	Creates a new account on the PChain under the username and password.
	Transfers funds from an XChain account owned by that username and password to the new PChain account.
	Returns the new, funded PChain account address.
*/
func (rpcManager RpcManager) transferAvaXChainToPChain(
	username string,
	password string,
	from string,
	to string,
	amount int) (string, error) {
	return "", nil
}