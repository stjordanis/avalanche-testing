package ethapi

import (
	"context"
	"fmt"
	"time"

	"math/big"

	"github.com/ava-labs/avalanche-testing/testsuite/tester"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	key                = "ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
	prefixedPrivateKey = fmt.Sprintf("PrivateKey-%s", key)
)

// NewEthAPIExecutor returns a new bombard test bombardExecutor
func NewEthAPIExecutor(ip string, port int, requestTimeout time.Duration) tester.AvalancheTester {
	return &ethAPIExecutor{
		ipAddr:         ip,
		port:           port,
		requestTimeout: requestTimeout,
	}
}

type ethAPIExecutor struct {
	ipAddr         string
	port           int
	requestTimeout time.Duration
}

// ExecuteTest implements the AvalancheTester interface
func (e *ethAPIExecutor) ExecuteTest() error {
	wsURI := fmt.Sprintf("ws://%s:%d/ext/bc/C/ws", e.ipAddr, e.port)
	client, err := ethclient.Dial(wsURI)
	if err != nil {
		return fmt.Errorf("Failed to create ethclient: %w", err)
	}
	logrus.Infof("Created ethclient")
	// avmClient := avm.NewClient(fmt.Sprintf("http://%s:%d"), "X", e.requestTimeout)
	// import private key
	// export AVAX
	// import AVAX to C Chain
	// verify balance
	//

	ctx := context.Background()

	cb58 := formatting.CB58{}
	factory := crypto.FactorySECP256K1R{}
	_ = cb58.FromString(key)
	pk, _ := factory.ToPrivateKey(cb58.Bytes)
	secpKey := pk.(*crypto.PrivateKeySECP256K1R)
	ethAddr := evm.GetEthAddress(secpKey)

	if err := testBasicAPICalls(ctx, client, ethAddr); err != nil {
		return fmt.Errorf("Basic API Calls failed: %w", err)
	}

	return nil
}

func testBasicAPICalls(ctx context.Context, client *ethclient.Client, ethAddr common.Address) error {
	headerChan := make(chan *types.Header)
	subscription, err := client.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		return fmt.Errorf("Failed to create subscription: %s", err)
	}
	logrus.Infof("Created subscription: %s", subscription)

	suggestedGasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get suggested gas price: %s", err)
	}
	logrus.Infof("Suggested gas price: %d", suggestedGasPrice.Uint64())

	logChan := make(chan types.Log)
	query := geth.FilterQuery{
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

	// Test Header and Block ByNumber work for special cases
	for i := 0; i > -3; i-- {
		if err := checkHeaderAndBlocks(ctx, client, i, ethAddr); err != nil {
			return err
		}
	}

	return nil
}

func checkHeaderAndBlocks(ctx context.Context, client *ethclient.Client, i int, ethAddr common.Address) error {
	header1, err := client.HeaderByNumber(ctx, big.NewInt(int64(i)))
	if err != nil {
		return fmt.Errorf("Failed to retrieve HeaderByNumber: %w", err)
	}
	logrus.Infof("HeaderByNumber (Block Number: %d, Block Hash: %s)", header1.Number, header1.Hash().Hex())

	originalHash := header1.Hash()
	originalBlockNumber := header1.Number
	if i >= 0 && int(originalBlockNumber.Int64()) != i {
		return fmt.Errorf("Requested block number %d, but found block number %d", i, originalBlockNumber)
	}

	header2, err := client.HeaderByHash(ctx, header1.Hash())
	if err != nil {
		return fmt.Errorf("Failed to retrieve HeaderByHash: %w", err)
	}
	logrus.Infof("HeaderByNumber (Block Number: %d, Block Hash: %s)", header2.Number, header2.Hash().Hex())

	if originalHash.Hex() != header2.Hash().Hex() || originalBlockNumber != header2.Number {
		return fmt.Errorf("Expected (Number, Hash) = (%s, %d), found (%s, %d)", originalHash.Hex(), originalBlockNumber, header2.Hash().Hex(), header2.Number)
	}

	block1, err := client.BlockByNumber(ctx, big.NewInt(int64(i)))
	if err != nil {
		return fmt.Errorf("Failed to retrieve BlockByNumber: %w", err)
	}
	header3 := block1.Header()
	logrus.Infof("BlockByNumber (Block Number: %d, Block Hash: %s)", header3.Number, header3.Hash().Hex())
	if originalHash.Hex() != header3.Hash().Hex() || originalBlockNumber != header3.Number {
		return fmt.Errorf("Expected (Number, Hash) = (%s, %d), found (%s, %d)", originalHash.Hex(), originalBlockNumber, header3.Hash().Hex(), header3.Number)
	}

	block2, err := client.BlockByHash(ctx, header1.Hash())
	if err != nil {
		return fmt.Errorf("Failed to retrieve BlockByHash: %w", err)
	}
	header4 := block2.Header()
	logrus.Infof("BlockByHash (Block Number: %d, Block Hash: %s)", header4.Number, header4.Hash().Hex())
	if originalHash.Hex() != header4.Hash().Hex() || originalBlockNumber != header4.Number {
		return fmt.Errorf("Expected (Number, Hash) = (%s, %d), found (%s, %d)", originalHash.Hex(), originalBlockNumber, header4.Hash().Hex(), header4.Number)
	}

	balance, err := client.BalanceAt(ctx, ethAddr, big.NewInt(int64(i)))
	if err != nil {
		return fmt.Errorf("Failed to get balance: %s", err)
	}
	logrus.Infof("Balance: %d for address: %s", balance, ethAddr.Hex())

	return nil
}

// Create functions to fund account
// Test Subscribe functionality
// Send a few transactions
// Deploy and call a smart contract
// Check on the API calls that have been broken in the past
// Atomic transactions with multicoin assets

// Check that all of the required APIs are enabled and working
// Then within kurtosis test, launch additional service and make
// sure that it is able to bootstrap the C Chain state

