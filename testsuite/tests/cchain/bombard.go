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
	ctx := context.Background()
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

	// Generate addresses
	addresses := p.numLists + 1
	pks := make([]*ecdsa.PrivateKey, addresses)
	addrs := make([]common.Address, addresses)
	for i := 0; i < addresses; i++ {
		pk, err := ethcrypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("problem creating new private key: %w", err)
		}
		ethAddr := ethcrypto.PubkeyToAddress(pk.PublicKey)
		pks[i] = pk
		addrs[i] = ethAddr
	}

	logrus.Infof("Moving all X-Chain Funds to C-Chain.")
	if err := workflowRunner.MoveBalanceToCChain(addrs[0]); err != nil {
		return err
	}

	amountPerAddress, txLimit, err := computeBalancePerAddress(ctx, ethClients[0], addrs[0], p.numLists)
	if err != nil {
		return err
	}

	if int64(p.numTxs) > txLimit.Int64() {
		return fmt.Errorf("want to send %d transactions per address but can only send %d", p.numTxs, txLimit)
	}

	logrus.Infof("Sending %d nAVAX each to %d C Chain addresses.", amountPerAddress, len(addrs)-1)
	if err := createAndConfirmTransfers(ctx, ethClients[0], pks[0], addrs[1:], 0, amountPerAddress); err != nil {
		return err
	}

	txLists := make([][]*types.Transaction, p.numLists)
	for i := 1; i < addresses; i++ {
		txs, err := createConsecutiveBasicEthTransactions(pks[i], addrs[i], 0, p.numTxs)
		if err != nil {
			return err
		}
		txLists[i-1] = txs
	}

	logrus.Infof("Issuing %d transactions.", p.numLists*p.numTxs)
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

	if err := confirmBlocks(context.Background(), finalizedHeight, ethClients); err != nil {
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
