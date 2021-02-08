package cchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/coreth/core/types"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/params"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

const (
	consecutiveHeights = 90
	waitForTipSleep    = time.Duration(1 * time.Second)
)

var (
	user               = api.UserPass{Username: "Jameson", Password: "Javier23r79h"}
	key                = "ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
	prefixedPrivateKey = fmt.Sprintf("PrivateKey-%s", key)
	avaxAmount         = uint64(10000000000000000)
	x2cConversion      = uint64(1000000000)
	cChainID           = big.NewInt(43112)
	signer             = types.NewEIP155Signer(cChainID)

	ethAddr common.Address
)

func init() {
	pkBytes, err := formatting.Decode(formatting.CB58, key)
	if err != nil {
		panic(err)
	}
	factory := crypto.FactorySECP256K1R{}
	pk, err := factory.ToPrivateKey(pkBytes)
	if err != nil {
		panic(err)
	}
	secpKey := pk.(*crypto.PrivateKeySECP256K1R)
	ethAddr = evm.GetEthAddress(secpKey)
}

// createConsecutiveBasicEthTransactions ...
func createConsecutiveBasicEthTransactions(pk *ecdsa.PrivateKey, addr common.Address, startingNonce uint64, numTxs int) ([]*types.Transaction, error) {
	txs := make([]*types.Transaction, numTxs)
	for i := 0; i < numTxs; i++ {
		nonce := uint64(i) + startingNonce
		tx := types.NewTransaction(nonce, addr, big.NewInt(1), 21000, big.NewInt(470*params.GWei), nil)
		signedTx, err := types.SignTx(tx, signer, pk)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %w", err)
		}
		txs[i] = signedTx
	}

	return txs, nil
}

func issueTxList(ctx context.Context, client *ethclient.Client, txs []*types.Transaction) error {
	for _, tx := range txs {
		if err := client.SendTransaction(ctx, tx); err != nil {
			return err
		}
	}

	return nil
}

func confirmTxList(ctx context.Context, client *ethclient.Client, txs []*types.Transaction) error {
	for _, tx := range txs {
		receipt, err := client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return fmt.Errorf("could not retrieve transaction %s: %w", tx.Hash().Hex(), err)
		}

		logrus.Infof("Transaction %s was in block: (%s, %d)", tx.Hash().Hex(), receipt.BlockHash.Hex(), receipt.BlockNumber)
	}

	return nil
}

// confirmBlocks ensures all *ethclient.Clients return the same blocks for
// a given height.
func confirmBlocks(ctx context.Context, maxHeight uint64, clients []*ethclient.Client) error {
	for i := uint64(0); i <= maxHeight; i++ {
		var hash string

		for j, c := range clients {
			b, err := c.BlockByNumber(ctx, big.NewInt(int64(i)))
			if err != nil {
				return err
			}

			blockHash := b.Hash().Hex()
			if len(hash) == 0 {
				hash = blockHash
				continue
			}

			if hash != blockHash {
				return fmt.Errorf("node %d got hash %s but expected %s for height %d", j, blockHash, hash, i)
			}
		}
	}

	return nil
}

// waitForStableTip ensures an array of *ethclient.Clients all report the same
// height before returning. If the clients return the same unequal heights for
// consecutiveHeights * waitForTipSleep, it is assumed that syncing is stalled
// and an error is returned.
func waitForStableTip(ctx context.Context, clients []*ethclient.Client) (uint64, time.Duration, error) {
	var (
		consecutiveSame int

		consecutiveDifferent int
		previousHeights      []uint64
	)

	for {
		var (
			heights   = make([]uint64, len(clients))
			foundDiff bool
		)

		for i, c := range clients {
			height, err := c.BlockNumber(ctx)
			if err != nil {
				return 0, 0, err
			}

			heights[i] = height
			if i != 0 && heights[0] != height {
				foundDiff = true
			}
		}

		if !foundDiff {
			consecutiveDifferent = 0
			previousHeights = nil

			consecutiveSame++
		} else {
			consecutiveSame = 0

			if len(previousHeights) > 0 && reflect.DeepEqual(previousHeights, heights) {
				consecutiveDifferent++
			} else {
				consecutiveDifferent = 1
				previousHeights = heights
			}
		}

		if consecutiveSame >= consecutiveHeights {
			return heights[0], time.Duration(consecutiveSame-1) * waitForTipSleep, nil
		}
		if consecutiveDifferent >= consecutiveHeights {
			return 0, 0, fmt.Errorf("block production is stuck at %v", heights)
		}

		time.Sleep(waitForTipSleep)
	}
}
