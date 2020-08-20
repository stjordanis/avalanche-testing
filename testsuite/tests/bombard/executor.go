package bombard

import (
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/avalanche-e2e-tests/testsuite/helpers"
	"github.com/ava-labs/avalanche-e2e-tests/testsuite/tester"
	"github.com/ava-labs/gecko/api"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// NewSendExecutor returns a new bombard test sendExecutor
func NewSendExecutor(clients []*apis.Client, numTxs, txFee uint64, acceptanceTimeout time.Duration) tester.AvalancheTester {
	return &sendExecutor{
		normalClients:     clients,
		numTxs:            numTxs,
		acceptanceTimeout: acceptanceTimeout,
		txFee:             txFee,
	}
}

type sendExecutor struct {
	fundedPrivateKey  string
	normalClients     []*apis.Client
	acceptanceTimeout time.Duration
	numTxs            uint64
	txFee             uint64
}

// ExecuteTest implements the AvalancheTester interface
func (e *sendExecutor) ExecuteTest() error {
	genesisClient := e.normalClients[0]
	secondaryClients := make([]*helpers.RPCWorkFlowRunner, len(e.normalClients)-1)
	xChainAddrs := make([]string, len(e.normalClients)-1)
	for i, client := range e.normalClients[1:] {
		secondaryClients[i] = helpers.NewRPCWorkFlowRunner(
			client,
			api.UserPass{Username: createRandomString(), Password: createRandomString()},
			e.acceptanceTimeout,
		)
		xChainAddress, _, err := secondaryClients[i].CreateDefaultAddresses()
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create default addresses for client: %d", i)
		}
		xChainAddrs[i] = xChainAddress
	}

	genesisUser := api.UserPass{Username: createRandomString(), Password: createRandomString()}
	highLevelGenesisClient := helpers.NewRPCWorkFlowRunner(
		genesisClient,
		genesisUser,
		e.acceptanceTimeout,
	)

	if _, err := highLevelGenesisClient.ImportGenesisFunds(); err != nil {
		return stacktrace.Propagate(err, "Failed to fund genesis client.")
	}
	addrs, err := genesisClient.XChainAPI().ListAddresses(genesisUser)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get genesis client's addresses")
	}
	if len(addrs) != 1 {
		return stacktrace.NewError("Found unexecpted number of addresses for genesis client: %d", len(addrs))
	}
	genesisAddress := addrs[0]
	logrus.Infof("Imported genesis funds at address: %s", genesisAddress)

	// Fund X Chain Addresses enough to issue numTxs
	seedAmount := e.numTxs * e.txFee
	if err := highLevelGenesisClient.FundXChainAddresses(xChainAddrs, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Failed to fund X Chain Addresses for Clients")
	}
	logrus.Infof("Funded X Chain Addresses.")

	for i, client := range secondaryClients {
		// Each address should have [e.txFee] remaining after sending [numTxs] and paying the fixed fee each time
		if err := client.VerifyXChainAVABalance(xChainAddrs[i], seedAmount); err != nil {
			return stacktrace.Propagate(err, "Failed to verify X Chain Balane for Client: %d", i)
		}
	}
	logrus.Infof("Verified X Chain Balances.")

	errChan := make(chan error, len(secondaryClients))
	for i, client := range secondaryClients {
		go client.SendAVAXBackAndForth(xChainAddrs[i], seedAmount, e.txFee, e.numTxs, errChan)
	}

	for i := 0; i < len(secondaryClients); i++ {
		err := <-errChan
		if err != nil {
			return err
		}
	}
	logrus.Infof("Finished sending back and forth for each client.")

	for i, client := range secondaryClients {
		// Each address should have [e.txFee] remaining after sending [numTxs] and paying the fixed fee each time
		if err := client.VerifyXChainAVABalance(xChainAddrs[i], e.txFee); err != nil {
			return stacktrace.Propagate(err, "Found unexpected X Chain Balane for Client: %d", i)
		}
	}

	logrus.Infof("Verified expected address balances.")

	return nil
	// txIDs := ids.Set{}
	// remainingAmount := seedAmount - uint64(e.numTxs+1)*e.txFee
	// for _, client := range secondaryClients {
	// 	txID, err := client.SendAVAX(genesisAddress, remainingAmount)
	// 	if err != nil {
	// 		return stacktrace.Propagate(err, "Failed to send AVAX back to genesis address.")
	// 	}
	// 	txIDs.Add(txID)
	// }

	// return highLevelGenesisClient.AwaitXChainTxs(txIDs.List()...)
}
