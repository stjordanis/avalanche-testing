package bombard

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche_client/apis"
	"github.com/ava-labs/avalanche-testing/testsuite/helpers"
	"github.com/ava-labs/avalanche-testing/testsuite/tester"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// NewBombardExecutor returns a new bombard test bombardExecutor
func NewBombardExecutor(clients []*apis.Client, numTxs, txFee uint64, acceptanceTimeout time.Duration, threadNum int) tester.AvalancheTester {
	return &bombardExecutor{
		normalClients:     clients,
		numTxs:            numTxs,
		acceptanceTimeout: acceptanceTimeout,
		txFee:             txFee,
		threadNum:	threadNum,
	}
}

type bombardExecutor struct {
	normalClients     []*apis.Client
	acceptanceTimeout time.Duration
	numTxs            uint64
	txFee             uint64
	threadNum	  int
}

func createRandomString() string {
	return fmt.Sprintf("rand:%d", rand.Int())
}

// ExecuteTest implements the AvalancheTester interface
func (e *bombardExecutor) ExecuteTest() error {
	logrus.Info("Bombard execution starts")
	genesisClient := e.normalClients[0]
	numSecondaryClients := len(e.normalClients)-1
	numReplicatedClients := e.threadNum
	secondaryClients := make([]*helpers.RPCWorkFlowRunner, numReplicatedClients)
	xChainAddrs := make([]string, numReplicatedClients)

	for j := 0; j < e.threadNum; j++ {
		clientIndex := j % numSecondaryClients
		client := e.normalClients[1 + clientIndex]
		secondaryClients[clientIndex] = helpers.NewRPCWorkFlowRunner(
			client,
			api.UserPass{Username: createRandomString(), Password: createRandomString()},
			e.acceptanceTimeout,
		)
		xChainAddress, _, err := secondaryClients[clientIndex].CreateDefaultAddresses()
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create default addresses for client: %d", j)
		}
		xChainAddrs[clientIndex] = xChainAddress
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

	// Fund X Chain Addresses enough to issue [numTxs]
	seedAmount := (e.numTxs + 1) * e.txFee
	if err := highLevelGenesisClient.FundXChainAddresses(xChainAddrs, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Failed to fund X Chain Addresses for Clients")
	}
	logrus.Infof("Funded X Chain Addresses with seedAmount %v.", seedAmount)
	logrus.Info("Will sleep 20 seconds... ")
	time.Sleep(20 * time.Second)
	codec, err := createXChainCodec()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to initialize codec.")
	}
	utxoLists := make([][]*avax.UTXO, len(secondaryClients))
	for i, client := range secondaryClients {
		// Each address should have [e.txFee] remaining after sending [numTxs] and paying the fixed fee each time
		if err := client.VerifyXChainAVABalance(xChainAddrs[i], seedAmount); err != nil {
			return stacktrace.Propagate(err, "Failed to verify X Chain Balane for Client: %d", i)
		}
		utxoReply, err := genesisClient.XChainAPI().GetUTXOs([]string{xChainAddrs[i]}, 10, "", "")
		if err != nil {
			return err
		}
		formattedUTXOs := utxoReply.UTXOs
		utxos := make([]*avax.UTXO, len(formattedUTXOs))
		for i, formattedUTXO := range formattedUTXOs {
			utxoBytes := formattedUTXO.Bytes
			utxo := &avax.UTXO{}
			err := codec.Unmarshal(utxoBytes, utxo)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to unmarshal utxo bytes.")
			}
			utxos[i] = utxo
		}
		utxoLists[i] = utxos
		logrus.Infof("Decoded %d UTXOs", len(utxos))

	}
	logrus.Infof("Verified X Chain Balances and retrieved UTXOs.")

	// Create a string of consecutive transactions for each secondary client to send
	privateKeys := make([]*crypto.PrivateKeySECP256K1R, len(secondaryClients))
	txLists := make([][][]byte, len(secondaryClients))
	txIDLists := make([][]ids.ID, len(secondaryClients))
	xChainID, err := e.normalClients[0].InfoAPI().GetBlockchainID("X")
	logrus.Info("X Chain ID ", xChainID)
	for j := 0; j < e.threadNum; j++ {
		clientIndex := j % numSecondaryClients
		client := e.normalClients[1 + clientIndex]
		nodeId, err := client.InfoAPI().GetNodeID()
		logrus.Info("Client with nodeID ", nodeId)
		utxo := utxoLists[clientIndex][0]
		pkStr, err := client.XChainAPI().ExportKey(secondaryClients[clientIndex].User(), xChainAddrs[clientIndex])
		if err != nil {
			return stacktrace.Propagate(err, "Failed to export key.")
		}

		if !strings.HasPrefix(pkStr, constants.SecretKeyPrefix) {
			return fmt.Errorf("private key missing %s prefix", constants.SecretKeyPrefix)
		}
		trimmedPrivateKey := strings.TrimPrefix(pkStr, constants.SecretKeyPrefix)
		formattedPrivateKey := formatting.CB58{}
		if err := formattedPrivateKey.FromString(trimmedPrivateKey); err != nil {
			return fmt.Errorf("problem parsing private key: %w", err)
		}

		factory := crypto.FactorySECP256K1R{}
		skIntf, err := factory.ToPrivateKey(formattedPrivateKey.Bytes)
		sk := skIntf.(*crypto.PrivateKeySECP256K1R)
		privateKeys[clientIndex] = sk

		logrus.Infof("Creating string of %d transactions", e.numTxs)
		txs, txIDs, err := CreateConsecutiveTransactions(utxo, e.numTxs, seedAmount, e.txFee, sk)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create transaction list.")
		}
		txLists[clientIndex] = txs
		txIDLists[clientIndex] = txIDs
	}
	logrus.Info("Tx list created .. .")
	wg := sync.WaitGroup{}
	issueTxsAsync := func(runner *helpers.RPCWorkFlowRunner, txList [][]byte) {
		if err = runner.IssueTxList(txList); err != nil {
			panic(err)
		}
		wg.Done()
	}
	numTotalTxs := uint64(len(txLists)) * e.numTxs
	startTime := time.Now()
	logrus.Info("Number of secondary clients ", len(secondaryClients))
	logrus.Info("Beginning to issue ", numTotalTxs, " transactions...")
	for i, client := range secondaryClients {
		wg.Add(1)
		go issueTxsAsync(client, txLists[i])
	}
	wg.Wait()

	duration := time.Since(startTime)
	logrus.Infof("Finished issuing transaction lists in %v seconds.", duration.Seconds())
	return nil
}
