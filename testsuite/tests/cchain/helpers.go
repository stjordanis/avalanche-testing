package cchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
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
	consecutiveHeights = 5
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
			return err
		}
		logrus.Infof("Transaction was in block: (%s, %d)", receipt.BlockHash.Hex(), receipt.BlockNumber)
	}

	return nil
}

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

func waitForStableTip(ctx context.Context, clients []*ethclient.Client) (uint64, error) {
	var consecutive int

	for {
		var (
			reportedHeight uint64
			foundDiff      bool
		)

		for _, c := range clients {
			height, err := c.BlockNumber(ctx)
			if err != nil {
				return 0, err
			}

			if reportedHeight == 0 {
				reportedHeight = height
				continue
			}

			if reportedHeight != height {
				foundDiff = true
				break
			}
		}

		if !foundDiff {
			consecutive++
		} else {
			consecutive = 0
		}

		if consecutive > consecutiveHeights {
			return reportedHeight, nil
		}

		time.Sleep(100 * time.Millisecond)
	}
}
