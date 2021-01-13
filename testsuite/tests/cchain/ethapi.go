package cchain

import (
	"context"
	"fmt"

	"math/big"

	"github.com/ava-labs/avalanche-testing/testsuite/tester"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/sirupsen/logrus"

	"github.com/ava-labs/coreth"
	"github.com/ava-labs/coreth/core/types"
	"github.com/ethereum/go-ethereum/common"
)

// NewEthAPIExecutor returns a new bombard test bombardExecutor
func NewEthAPIExecutor(client *ethclient.Client) tester.AvalancheTester {
	return &ethAPIExecutor{
		client: client,
	}
}

type ethAPIExecutor struct {
	client *ethclient.Client
}

// ExecuteTest implements the AvalancheTester interface
func (e *ethAPIExecutor) ExecuteTest() error {
	ctx := context.Background()

	logrus.Info("Conducting test on basic ethclient API calls")
	if err := testBasicAPICalls(ctx, e.client, ethAddr); err != nil {
		return fmt.Errorf("Basic API Calls failed: %w", err)
	}
	logrus.Info("Basic API Call test was successful.")

	return nil
}

// testBasicAPICalls ...
func testBasicAPICalls(ctx context.Context, client *ethclient.Client, ethAddr common.Address) error {
	if err := testFetchHeadersAndBlocks(ctx, client, ethAddr); err != nil {
		return err
	}

	if err := testSuggestGasPrice(ctx, client); err != nil {
		return err
	}

	if err := testSubscription(ctx, client); err != nil {
		return err
	}

	return nil
}

// testSuggestGasPrice ...
func testSuggestGasPrice(ctx context.Context, client *ethclient.Client) error {
	suggestedGasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get suggested gas price: %s", err)
	}
	logrus.Infof("Suggested gas price: %d", suggestedGasPrice.Uint64())
	return nil
}

// testSubscription ...
func testSubscription(ctx context.Context, client *ethclient.Client) error {
	headerChan := make(chan *types.Header)
	subscription, err := client.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		return fmt.Errorf("Failed to create subscription: %s", err)
	}
	logrus.Infof("Created subscription: %s", subscription)

	logChan := make(chan types.Log)
	query := coreth.FilterQuery{
		BlockHash: nil,
		FromBlock: nil,
		ToBlock:   nil,
		Addresses: []common.Address{},
		Topics:    [][]common.Hash{},
	}
	subscription, err = client.SubscribeFilterLogs(ctx, query, logChan)
	if err != nil {
		return fmt.Errorf("Failed to create subscription: %s", err)
	}
	logrus.Infof("Created subscription: %s", subscription)

	return nil
}

// testFetchHeadersAndBlocks ...
func testFetchHeadersAndBlocks(ctx context.Context, client *ethclient.Client, ethAddr common.Address) error {
	// Test Header and Block ByNumber work for special cases
	for i := 0; i > -3; i-- {
		if err := checkHeaderAndBlocks(ctx, client, i, ethAddr); err != nil {
			return err
		}
	}

	return nil
}

// checkHeaderAndBlocks ...
func checkHeaderAndBlocks(ctx context.Context, client *ethclient.Client, i int, ethAddr common.Address) error {
	header1, err := client.HeaderByNumber(ctx, big.NewInt(int64(i)))
	if err != nil {
		return fmt.Errorf("Failed to retrieve HeaderByNumber: %w", err)
	}
	logrus.Infof("HeaderByNumber (Block Number: %d, Block Hash: %s)", header1.Number, header1.Hash().Hex())

	originalHash := header1.Hash()
	originalBlockNumber := header1.Number
	if i >= 0 && int(originalBlockNumber.Int64()) != i {
		return fmt.Errorf("requested block number %d, but found block number %d", i, originalBlockNumber)
	}

	header2, err := client.HeaderByHash(ctx, header1.Hash())
	if err != nil {
		return fmt.Errorf("Failed to retrieve HeaderByHash: %w", err)
	}
	logrus.Infof("HeaderByNumber (Block Number: %d, Block Hash: %s)", header2.Number, header2.Hash().Hex())

	if originalHash.Hex() != header2.Hash().Hex() {
		return fmt.Errorf("expected HeaderByNumber to return block with hash %s, but found %s", originalHash.Hex(), header2.Hash().Hex())
	}

	if originalBlockNumber.Cmp(header2.Number) != 0 {
		return fmt.Errorf("expected HeaderByNumber to return block with number %d, but found %d", originalBlockNumber, header2.Number)
	}

	block1, err := client.BlockByNumber(ctx, big.NewInt(int64(i)))
	if err != nil {
		return fmt.Errorf("Failed to retrieve BlockByNumber: %w", err)
	}
	header3 := block1.Header()
	logrus.Infof("BlockByNumber (Block Number: %d, Block Hash: %s)", header3.Number, header3.Hash().Hex())

	if originalHash.Hex() != header3.Hash().Hex() {
		return fmt.Errorf("expected BlockByNumber to return block with hash %s, but found %s", originalHash.Hex(), header3.Hash().Hex())
	}

	if originalBlockNumber.Cmp(header3.Number) != 0 {
		return fmt.Errorf("expected BlockByNumber to return block with number %d, but found %d", originalBlockNumber, header3.Number)
	}

	block2, err := client.BlockByHash(ctx, header1.Hash())
	if err != nil {
		return fmt.Errorf("Failed to retrieve BlockByHash: %w", err)
	}
	header4 := block2.Header()
	logrus.Infof("BlockByHash (Block Number: %d, Block Hash: %s)", header4.Number, header4.Hash().Hex())

	if originalHash.Hex() != header4.Hash().Hex() {
		return fmt.Errorf("expected BlockByHash to return block with hash %s, but found %s", originalHash.Hex(), header4.Hash().Hex())
	}

	if originalBlockNumber.Cmp(header4.Number) != 0 {
		return fmt.Errorf("expected BlockByHash to return block with number %d, but found %d", originalBlockNumber, header4.Number)
	}

	balance, err := client.BalanceAt(ctx, ethAddr, big.NewInt(int64(i)))
	if err != nil {
		return fmt.Errorf("Failed to get balance: %s", err)
	}
	logrus.Infof("Balance: %d for address: %s", balance, ethAddr.Hex())

	return nil
}
