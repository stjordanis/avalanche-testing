package cchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

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

func confirmBlocks(ctx context.Context, clients []*ethclient.Client) error {
	i := uint64(0)
	marker := clients[0]
	for {
		height, err := marker.BlockNumber(ctx)
		if err != nil {
			return err
		}

		if i >= height {
			return nil
		}
		logrus.Infof("Checking Block: %d", height)

		var hash string
		for j, c := range clients {
			b, err := c.BlockByNumber(ctx, big.NewInt(int64(i)))
			if err != nil {
				return err
			}

			if len(hash) == 0 {
				hash = b.Hash().Hex()
				continue
			}

			if hash != b.Hash().Hex() {
				return fmt.Errorf("node %d got hash %s but expected %s for height %d", j, b.Hash().Hex(), hash, i)
			}
		}

		i++
	}
}
