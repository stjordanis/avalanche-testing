package cchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/avm"
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
	requestTimeout     = 3 * time.Second
	avaxAmount         = uint64(10000000000000000)
	x2cConversion      = uint64(1000000000)
	cChainID           = big.NewInt(43112)
	signer             = types.NewEIP155Signer(cChainID)

	ethAddr common.Address
)

func init() {
	cb58 := formatting.CB58{}
	factory := crypto.FactorySECP256K1R{}
	err := cb58.FromString(key)
	if err != nil {
		panic(err)
	}
	pk, err := factory.ToPrivateKey(cb58.Bytes)
	if err != nil {
		panic(err)
	}
	secpKey := pk.(*crypto.PrivateKeySECP256K1R)
	ethAddr = evm.GetEthAddress(secpKey)
}

func confirmTx(c *avm.Client, txID ids.ID) error {
	for {
		status, err := c.GetTxStatus(txID)
		if err != nil {
			return err
		}

		if status == choices.Accepted {
			return nil
		}

		logrus.Infof("Status of %s was %s", txID, status)
		time.Sleep(time.Second)
	}
}

// createConsecutiveBasicEthTransactions ...
func createConsecutiveBasicEthTransactions(pk *ecdsa.PrivateKey, addr common.Address, startingNonce uint64, numTxs uint64) ([]*types.Transaction, error) {
	txs := make([]*types.Transaction, numTxs)
	for i := startingNonce; i < numTxs; i++ {
		nonce := i + startingNonce
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

// fundCChainAddress ...
func fundCChainAddress(avmClient *avm.Client, cChainClient *evm.Client, addr common.Address, avaxAmount uint64) error {
	xAddr, err := avmClient.ImportKey(user, prefixedPrivateKey)
	if err != nil {
		panic(err)
	}

	_, err = cChainClient.ImportKey(user, prefixedPrivateKey)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	cChainBech32 := fmt.Sprintf("C%s", xAddr[1:])
	txID, err := avmClient.ExportAVAX(user, nil, "", avaxAmount, cChainBech32)
	if err != nil {
		return fmt.Errorf("Failed to export AVAX to C-Chain: %w", err)
	}
	logrus.Infof("Exported %d AVAX to %s. TxID: %s", avaxAmount, cChainBech32, txID)
	confirmTx(avmClient, txID)

	txID, err = cChainClient.Import(user, addr.Hex(), "X")
	if err != nil {
		return fmt.Errorf("Failed to import AVAX to C-Chain: %w", err)
	}
	logrus.Infof("Imported AVAX to %s. TxID: %s", addr.Hex(), txID)

	time.Sleep(2 * time.Second)

	cBalance, err := ethClient.BalanceAt(ctx, addr, nil)
	if err != nil {
		return fmt.Errorf("Failed to get balance of %s on C Chain: %w", addr, err)
	}
	logrus.Infof("Found balance of %d", cBalance)

	cAmount := new(big.Int).Mul(big.NewInt(int64(avaxAmount)), big.NewInt(int64(x2cConversion)))
	if cAmount.Cmp(cBalance) != 0 {
		return fmt.Errorf("Found unexpected balance: %d, expected %d", cBalance, cAmount)
	}

	return nil
}

// fundRandomCChainAddresses generates [num] private keys and imports [amount] AVAX (not nAVAX) to
// each of the generated keys
func fundRandomCChainAddresses(avmClient *avm.Client, cChainClient *evm.Client, num int, amount uint64) ([]*ecdsa.PrivateKey, []common.Address, error) {
	pks := make([]*ecdsa.PrivateKey, num)
	addrs := make([]common.Address, num)
	for i := 0; i < num; i++ {
		pk, err := ethcrypto.GenerateKey()
		if err != nil {
			return nil, nil, fmt.Errorf("problem creating new private key: %w", err)
		}
		ethAddr := ethcrypto.PubkeyToAddress(pk.PublicKey)
		pks[i] = pk
		addrs[i] = ethAddr

		if err := fundCChainAddress(avmClient, cChainClient, ethAddr, amount); err != nil {
			return nil, nil, err
		}
	}

	return pks, addrs, nil
}
