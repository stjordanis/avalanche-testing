package cchain

import (
	"context"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/testsuite/tester"
	"github.com/ava-labs/coreth/core/types"
	"github.com/ava-labs/coreth/ethclient"
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
	xClient := p.client.XChainAPI()
	cClient := p.client.CChainAPI()
	cEthClient := p.client.CChainEthAPI()

	pks, addrs, err := fundRandomCChainAddresses(xClient, cClient, cEthClient, p.numLists, avaxAmount)
	if err != nil {
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
