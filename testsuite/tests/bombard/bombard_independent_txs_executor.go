package bombard

import (
	"fmt"
	"sync"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/testsuite/helpers"
	"github.com/ava-labs/avalanche-testing/testsuite/tester"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/crypto"
	cjson "github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// NewBombardIndependentTxsExecutor returns a new bombardIndependentTxsExecutor
func NewBombardIndependentTxsExecutor(numTxs, txFee uint64, acceptanceTimeout time.Duration) tester.AvalancheConfigurableTester {
	return &bombardIndependentTxsExecutor{
		numTxs:            numTxs,
		acceptanceTimeout: acceptanceTimeout,
		txFee:             txFee,
	}
}

type bombardIndependentTxsExecutor struct {
	normalClients     []*services.Client
	acceptanceTimeout time.Duration
	numTxs            uint64
	txFee             uint64
}

func (e *bombardIndependentTxsExecutor) SetClients(clients []*services.Client) error {
	if len(clients) < 2 {
		return fmt.Errorf("bombard executor requires at least two clients, but was only given %d", len(clients))
	}
	e.normalClients = clients
	return nil
}

// ExecuteTest implements the AvalancheTester interface
func (e *bombardIndependentTxsExecutor) ExecuteTest() error {
	secondaryClients := make([]*helpers.RPCWorkFlowRunner, len(e.normalClients))
	xChainAddrs := make([]string, len(e.normalClients))
	for i, client := range e.normalClients {
		secondaryClients[i] = helpers.NewRPCWorkFlowRunner(
			client,
			api.UserPass{Username: helpers.CreateRandomString(), Password: helpers.CreateRandomString()},
			e.acceptanceTimeout,
		)
		xChainAddress, _, err := secondaryClients[i].CreateDefaultAddresses()
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create default addresses for client: %d", i)
		}
		xChainAddrs[i] = xChainAddress
	}

	genesisClient := e.normalClients[0]
	highLevelGenesisClient := helpers.NewRPCWorkFlowRunner(
		genesisClient,
		api.UserPass{Username: helpers.CreateRandomString(), Password: helpers.CreateRandomString()},
		e.acceptanceTimeout,
	)
	genesisAddress, err := highLevelGenesisClient.ImportGenesisFunds()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to fund genesis client.")
	}
	logrus.Infof("Imported genesis funds at address: %s", genesisAddress)

	// Fund X Chain Addresses enough to issue [numTxs]
	seedAmount := (e.numTxs+1)*e.txFee + e.numTxs
	if err := highLevelGenesisClient.FundXChainAddresses(xChainAddrs, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Failed to fund X Chain Addresses for Clients")
	}
	logrus.Infof("Funded X Chain Addresses with seedAmount %v.", seedAmount)

	codec, err := helpers.CreateXChainCodec()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to initialize codec.")
	}
	for i, client := range secondaryClients {
		// Each address should have [e.txFee] remaining after sending [numTxs] and paying the fixed fee each time
		if err := client.VerifyXChainAVABalance(xChainAddrs[i], seedAmount); err != nil {
			return stacktrace.Propagate(err, "Failed to verify X Chain Balane for Client: %d", i)
		}
	}
	logrus.Infof("Verified X Chain Balances and retrieved UTXOs.")

	// Create a string of consecutive transactions for each secondary client to send
	privateKeys := make([]*crypto.PrivateKeySECP256K1R, len(secondaryClients))
	txLists := make([][][]byte, len(secondaryClients))
	txIDLists := make([][]ids.ID, len(secondaryClients))
	for i, client := range e.normalClients {
		xChainClient := client.XChainAPI()
		pkStr, err := xChainClient.ExportKey(secondaryClients[i].User(), xChainAddrs[i])
		if err != nil {
			return stacktrace.Propagate(err, "Failed to export key.")
		}

		sk, err := helpers.ConvertFormattedPrivateKey(pkStr)
		if err != nil {
			return stacktrace.Propagate(err, "failed to convert private key for client %d", i)
		}
		privateKeys[i] = sk

		// If the transactions should be independent create a transaction with [e.numTxs] outputs
		// to create transactions that are not chained
		sendOutputs := make([]avm.SendOutput, 0, e.numTxs)
		originalUTXOAmount := e.txFee + 1
		for j := 0; j < int(e.numTxs); j++ {
			sendOutputs = append(sendOutputs, avm.SendOutput{
				AssetID: "AVAX",
				Amount:  cjson.Uint64(originalUTXOAmount),
				To:      xChainAddrs[i],
			})
		}
		txID, err := xChainClient.SendMultiple(secondaryClients[i].User(), nil, "", sendOutputs, "")
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create UTXOs for independent transactions.")
		}
		status, _, err := xChainClient.ConfirmTx(txID, 5, time.Second)
		if err != nil {
			return err
		}
		if status != choices.Accepted {
			return stacktrace.NewError("Failed to confirm create UTXOs via SendMultiple transaction %s for client %d. Status: %s.", txID, i, status)
		}

		utxosBytes := make([][]byte, 0, e.numTxs)
		utxoBytesIter, index, err := xChainClient.GetUTXOs([]string{xChainAddrs[i]}, 0, "", "")
		for {
			if err != nil {
				return stacktrace.Propagate(err, "Failed to get UTXOs")
			}
			if len(utxoBytesIter) == 0 {
				break
			}
			utxosBytes = append(utxosBytes, utxoBytesIter...)
			utxoBytesIter, index, err = xChainClient.GetUTXOs([]string{xChainAddrs[i]}, 0, index.Address, index.UTXO)
		}

		if len(utxosBytes) != int(e.numTxs) {
			return stacktrace.NewError("Found unexpected number of UTXOs %d. Expected to find %d", len(utxosBytes), e.numTxs)
		}

		utxos := make([]*avax.UTXO, len(utxosBytes))
		for i, utxoBytes := range utxosBytes {
			utxo := &avax.UTXO{}
			_, err := codec.Unmarshal(utxoBytes, utxo)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to unmarshal utxo bytes.")
			}
			utxos[i] = utxo
		}

		logrus.Infof("Creating string of %d independent transactions", e.numTxs)
		txs, txIDs, err := helpers.CreateIndependentBurnTxs(utxos, originalUTXOAmount, e.txFee, sk)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create transaction list.")
		}
		txLists[i] = txs
		txIDLists[i] = txIDs
	}

	wg := sync.WaitGroup{}
	issueTxsAsync := func(runner *helpers.RPCWorkFlowRunner, txList [][]byte) {
		defer wg.Done()

		if err := runner.IssueTxList(txList); err != nil {
			panic(err)
		}
	}

	startTime := time.Now()
	logrus.Infof("Beginning to issue transactions...")
	for i, client := range secondaryClients {
		wg.Add(1)
		issueTxsAsync(client, txLists[i])
	}
	wg.Wait()

	duration := time.Since(startTime)
	logrus.Infof("Finished issuing transaction lists in %v seconds.", duration.Seconds())

	allTxIDs := make([]ids.ID, 0)
	for _, txIDs := range txIDLists {
		allTxIDs = append(allTxIDs, txIDs...)
	}
	txIDsSet := ids.Set{}
	txIDsSet.Add(allTxIDs...)
	if txIDsSet.Len() != len(allTxIDs) {
		return fmt.Errorf("expected set of txIDs to have the same size as the issued transactions list, but was %d. Expected %d", txIDsSet.Len(), len(allTxIDs))
	}

	for i, client := range secondaryClients {
		logrus.Infof("Confirming %d txs on client %d", len(allTxIDs), i)
		if err := client.AwaitXChainTxs(allTxIDs); err != nil {
			return stacktrace.Propagate(err, "Failed to confirm transactions for client %d.", i)
		}
	}
	logrus.Infof("Confirmed %d transactions on every client.", len(allTxIDs))

	return nil
}
