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
	// Create a genesis keystore user for interacting with the genesis funds
	// without importing it to any other client
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

	originalUTXOAmount := e.txFee + 1

	// Create secondary clients on each node and create and export a private key on each
	// one, fund the corresponding address, and use the genesis client to fund it with sufficient
	// UTXOs for the test.
	secondaryClients := make([]*helpers.RPCWorkFlowRunner, len(e.normalClients))
	codec, err := helpers.CreateXChainCodec()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to initialize codec.")
	}

	txLists := make([][][]byte, len(secondaryClients))
	txIDLists := make([][]ids.ID, len(secondaryClients))

	for i, client := range e.normalClients {
		secondaryClient := helpers.NewRPCWorkFlowRunner(
			client,
			api.UserPass{Username: helpers.CreateRandomString(), Password: helpers.CreateRandomString()},
			e.acceptanceTimeout,
		)
		secondaryClients[i] = secondaryClient
		xChainAddress, _, err := secondaryClient.CreateDefaultAddresses()
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create default addresses for client: %d", i)
		}

		// Export formatted private key and convert it to a private key type
		secondaryClientXChainAPI := secondaryClient.Client().XChainAPI()
		pkStr, err := secondaryClientXChainAPI.ExportKey(secondaryClients[i].User(), xChainAddress)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to export key.")
		}

		sk, err := helpers.ConvertFormattedPrivateKey(pkStr)
		if err != nil {
			return stacktrace.Propagate(err, "failed to convert private key for client %d", i)
		}

		numOutputs := e.numTxs
		if numOutputs > 2000 {
			numOutputs = 2000
		}
		sendOutputs := make([]avm.SendOutput, 0, numOutputs)
		for j := uint64(0); j < numOutputs; j++ {
			sendOutputs = append(sendOutputs, avm.SendOutput{
				AssetID: "AVAX",
				Amount:  cjson.Uint64(originalUTXOAmount),
				To:      xChainAddress,
			})
		}

		numRemaining := e.numTxs
		for numRemaining > 0 {
			if len(sendOutputs) > int(numRemaining) {
				sendOutputs = sendOutputs[:int(numRemaining)]
			}
			numRemaining -= uint64(len(sendOutputs))
			txID, err := genesisClient.XChainAPI().SendMultiple(highLevelGenesisClient.User(), nil, "", sendOutputs, "")
			if err != nil {
				return stacktrace.Propagate(err, "Failed to send transaction with %d outputs", len(sendOutputs))
			}
			logrus.Infof("Sent transaction %s with %d outputs", txID, len(sendOutputs))

			if err := highLevelGenesisClient.AwaitXChainTransactionAcceptance(txID); err != nil {
				return stacktrace.Propagate(err, "Failed to confirm transaction %s", txID)
			}
		}

		utxosBytes := make([][]byte, 0, e.numTxs)
		utxoBytesIter, index, err := secondaryClientXChainAPI.GetUTXOs([]string{xChainAddress}, 0, "", "")
		for {
			if err != nil {
				return stacktrace.Propagate(err, "Failed to get UTXOs")
			}
			if len(utxoBytesIter) == 0 {
				break
			}
			utxosBytes = append(utxosBytes, utxoBytesIter...)
			utxoBytesIter, index, err = secondaryClientXChainAPI.GetUTXOs([]string{xChainAddress}, 0, index.Address, index.UTXO)
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
