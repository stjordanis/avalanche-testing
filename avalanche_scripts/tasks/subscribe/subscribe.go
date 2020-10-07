package subscribe

import (
	"context"
	"math/big"

	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

// DoTask ...
func DoTask() {
	client, err := ethclient.Dial("ws://127.0.0.1:9650/ext/bc/C/ws")
	if err != nil {
		logrus.Errorf("Failed to create client: %s", err)
		return
	}
	logrus.Infof("Created client")

	ctx := context.Background()

	key := "ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
	cb58 := formatting.CB58{}
	factory := crypto.FactorySECP256K1R{}
	_ = cb58.FromString(key)
	pk, _ := factory.ToPrivateKey(cb58.Bytes)
	secpKey := pk.(*crypto.PrivateKeySECP256K1R)
	ethAddr := evm.GetEthAddress(secpKey)
	// shortIDAddr := pk.PublicKey().Address()

	// headerChan := make(chan *types.Header)
	// subscription, err := client.SubscribeNewHead(ctx, headerChan)
	// if err != nil {
	// 	logrus.Errorf("Failed to create subscription: %s", err)
	// 	return
	// }
	// logrus.Infof("Created subscription: %s", subscription)

	// suggestedGasPrice, err := client.SuggestGasPrice(ctx)
	// if err != nil {
	// 	logrus.Errorf("Failed to get suggested gas price: %s", err)
	// 	return
	// }
	// logrus.Infof("Suggested gas price: %d", suggestedGasPrice.Uint64())

	// logChan := make(chan types.Log)
	// query := geth.FilterQuery{
	// 	BlockHash: nil,
	// 	FromBlock: nil,
	// 	ToBlock:   nil,
	// 	Addresses: []common.Address{},
	// 	Topics:    [][]common.Hash{},
	// }
	// subscription, err := client.SubscribeFilterLogs(ctx, query, logChan)
	// if err != nil {
	// 	logrus.Errorf("Failed to create subscription: %s", err)
	// 	return
	// }
	// logrus.Infof("Created subscription: %s", subscription)

	// for i := -1; i < 3; i++ {
	// 	header, err := client.HeaderByNumber(ctx, big.NewInt(int64(i)))
	// 	if err != nil {
	// 		logrus.Errorf("err: %s", err)
	// 		return
	// 	}
	// 	logrus.Infof("Header: %s", header)
	// 	block, err := client.BlockByNumber(ctx, big.NewInt(int64(i)))
	// 	if err != nil {
	// 		logrus.Errorf("err: %s", err)
	// 		return
	// 	}
	// 	logrus.Infof("Block: %s", block)
	// }

	// blockNrOrHash := rpc.BlockNumberOrHash{
	// 	BlockNumber: int64(2)
	// }

	balance, err := client.BalanceAt(ctx, ethAddr, big.NewInt(int64(1)))
	if err != nil {
		logrus.Errorf("Failed to get balance: %s", err)
		return
	}
	logrus.Infof("Balance: %d for address: %s", balance, ethAddr.Hex())

}
