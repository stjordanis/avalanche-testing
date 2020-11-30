package cchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/testsuite/helpers"
	"github.com/ava-labs/avalanche-testing/testsuite/tester"
	"github.com/ava-labs/coreth/core/types"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/sirupsen/logrus"
)

type parallelBasicTxXputTest struct {
	client   *services.Client
	numLists int
	numTxs   int
}

// NewBasicTransactionThroughputTest returns a test executor that will run a small xput test of [numTxs] from each of [numLists] accounts
// Note: all issued to the same node.
func NewBasicTransactionThroughputTest(client *services.Client, numLists int, numTxs int) tester.AvalancheTester {
	return &parallelBasicTxXputTest{
		client:   client,
		numLists: numLists,
		numTxs:   numTxs,
	}
}

// ExecuteTest ...
func (p *parallelBasicTxXputTest) ExecuteTest() error {
	workflowRunner := helpers.NewRPCWorkFlowRunner(
		p.client,
		user,
		3*time.Second,
	)
	cEthClient := p.client.CChainEthAPI()

	pks := make([]*ecdsa.PrivateKey, p.numLists)
	addrs := make([]common.Address, p.numLists)
	for i := 0; i < p.numLists; i++ {
		pk, err := ethcrypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("problem creating new private key: %w", err)
		}
		ethAddr := ethcrypto.PubkeyToAddress(pk.PublicKey)
		pks[i] = pk
		addrs[i] = ethAddr
	}

	logrus.Infof("Funding %d C Chain addresses.", len(addrs))
	if err := workflowRunner.FundCChainAddresses(addrs, avaxAmount); err != nil {
		return err
	}

	txLists := make([][]*types.Transaction, p.numLists)
	for i := 0; i < p.numLists; i++ {
		txs, err := createConsecutiveBasicEthTransactions(pks[i], addrs[i], 0, p.numTxs)
		if err != nil {
			return err
		}
		txLists[i] = txs
	}

	errs := make(chan error, p.numLists)
	launchIssueTxList := func(ctx context.Context, ethclient *ethclient.Client, txList []*types.Transaction) {
		err := issueTxList(ctx, cEthClient, txList)
		errs <- err
	}
	launchedIssuers := time.Now()
	for _, txList := range txLists {
		go launchIssueTxList(context.Background(), cEthClient, txList)
	}
	startedGoRoutines := time.Now()
	logrus.Infof("Took %v to launch issuers", startedGoRoutines.Sub(launchedIssuers).Seconds())

	for i := 0; i < p.numLists; i++ {
		err := <-errs
		if err != nil {
			panic(err)
		}
	}
	finishedIssuing := time.Now()
	logrus.Infof("Took %v to finish issuing after launching issuers", finishedIssuing.Sub(launchedIssuers).Seconds())

	time.Sleep(3 * time.Second)
	for _, txList := range txLists {
		if err := confirmTxList(context.Background(), cEthClient, txList); err != nil {
			return err
		}
	}
	finishedConfirming := time.Now()
	logrus.Infof("Finished confirming after %v seconds. Total time start to finish: %v", finishedConfirming.Sub(finishedIssuing).Seconds(), finishedConfirming.Sub(launchedIssuers).Seconds())

	return nil
}
