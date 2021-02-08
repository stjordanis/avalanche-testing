package cchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"
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

type parallelTxXputTest struct {
	client   []*services.Client
	numLists int
	numTxs   int
}

// NewBasicTransactionThroughputTest returns a test executor that will run
// a small xput test of [numTxs] from each of [numLists] accounts
//
// Note: all transactions issued to the same node.
func NewBasicTransactionThroughputTest(client *services.Client, numLists int, numTxs int) tester.AvalancheTester {
	return &parallelTxXputTest{
		client:   []*services.Client{client},
		numLists: numLists,
		numTxs:   numTxs,
	}
}

// NewContentiousBlockThroughputTest returns a test executor that will run
// a xput test of [numTxs] from each of [numLists] accounts
//
// Note: transactions for a given account will be issued from the same node but
// transactions for different accounts could be broadcast from different nodes.
func NewContentiousBlockThroughputTest(clients []*services.Client, numLists int, numTxs int) tester.AvalancheTester {
	return &parallelTxXputTest{
		client:   clients,
		numLists: numLists,
		numTxs:   numTxs,
	}
}

// ExecuteTest ...
func (p *parallelTxXputTest) ExecuteTest() error {
	// create first client that funds rest of clients
	// import funds to all addresses at start of test
	funder := p.client[0]
	workflowRunner := helpers.NewRPCWorkFlowRunner(
		funder,
		user,
		3*time.Second,
	)

	ethClients := make([]*ethclient.Client, len(p.client))
	for i, c := range p.client {
		ethClients[i] = c.CChainEthAPI()
	}

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
		err := issueTxList(ctx, ethclient, txList)
		errs <- err
	}
	launchedIssuers := time.Now()
	launchClient := 0
	for _, txList := range txLists {
		go launchIssueTxList(context.Background(), ethClients[launchClient], txList)

		launchClient++
		if launchClient >= len(ethClients) {
			launchClient = 0
		}
	}
	startedGoRoutines := time.Now()
	logrus.Infof("Took %v to launch issuers", startedGoRoutines.Sub(launchedIssuers))

	var (
		wg       sync.WaitGroup
		groupErr error

		finalizedHeight   uint64
		stableTip         time.Time
		stableTipDuration time.Duration
	)

	wg.Add(2)
	go func() {
		height, waiting, err := waitForStableTip(context.Background(), ethClients)
		if err != nil {
			logrus.Error(err)
			groupErr = err
		}
		finalizedHeight = height
		stableTip = time.Now()
		stableTipDuration = stableTip.Sub(launchedIssuers) - waiting
		logrus.Infof("Took %v to reach stable tip", stableTipDuration)
		wg.Done()
	}()

	go func() {
		for i := 0; i < p.numLists; i++ {
			err := <-errs
			if err != nil {
				logrus.Error(err)
				groupErr = err
			}
		}
		finishedIssuing := time.Now()
		logrus.Infof("Took %v to finish issuing (after launching issuers)", finishedIssuing.Sub(launchedIssuers))
		wg.Done()
	}()

	wg.Wait()
	if groupErr != nil {
		return groupErr
	}

	err := confirmBlocks(context.Background(), finalizedHeight, ethClients)
	if err != nil {
		return err
	}

	finishedVerifying := time.Now()
	logrus.Infof("Took %v to verify blocks", finishedVerifying.Sub(stableTip))

	for _, txList := range txLists {
		if err := confirmTxList(context.Background(), funder.CChainEthAPI(), txList); err != nil {
			return err
		}
	}

	finishedConfirming := time.Now()
	logrus.Infof("Took %v to confirm txs", finishedConfirming.Sub(finishedVerifying))

	totalTxs := p.numLists * p.numTxs
	tps := float64(totalTxs) / stableTipDuration.Seconds()
	logrus.Infof(
		"Finalized %d transactions in %v (%f TPS). Checking time start to finish: %v",
		totalTxs,
		stableTipDuration,
		tps,
		finishedConfirming.Sub(launchedIssuers),
	)

	return nil
}
