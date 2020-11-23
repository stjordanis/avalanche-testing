package cchain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche_client/apis"
	cjson "github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// atomicWorkflowTest tests the import/export flow of both
// AVAX (native coin) and ANTs (native tokens) between the
// X and C Chains
type atomicWorkflowTest struct {
	client *apis.Client
	txFee  uint64
}

// CreateAtomicWorkflowTest returns a test of import/export transactions between X <-> C
func CreateAtomicWorkflowTest(client *apis.Client, txFee uint64) avalanche.Tester {
	return &atomicWorkflowTest{
		client: client,
		txFee:  txFee,
	}
}

func (aw *atomicWorkflowTest) ExecuteTest() error {
	logrus.Infof("Executing atomic workflow test")
	_, _ = aw.client.KeystoreAPI().CreateUser(user)
	xClient := aw.client.XChainAPI()
	cClient := aw.client.CChainAPI()
	cEthClient := aw.client.CChainEthAPI()

	xAddr, err := xClient.ImportKey(user, prefixedPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to import private key to X Chain: %w", err)
	}
	balanceReply, err := xClient.GetBalance(xAddr, "AVAX")
	if err != nil {
		return fmt.Errorf("failed to get AVAX balance: %w", err)
	}

	cAddr, err := cClient.ImportKey(user, prefixedPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to import private key to C Chain: %w", err)
	}

	expectedAVAXBalance := uint64(balanceReply.Balance)
	exportAVAXAmount := uint64(10000000000)
	assetAmount := uint64(100000000)
	exportAssetAmount := assetAmount - 10
	expectedAssetBalance := assetAmount
	logrus.Infof("Created clients and retrieved initial balance of %d", expectedAVAXBalance)

	logrus.Infof("Creating new asset")
	assetID, err := xClient.CreateAsset(user, nil, "", "TestToken", "TEST", 1, []*avm.Holder{
		{
			Amount:  cjson.Uint64(assetAmount),
			Address: xAddr,
		},
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}
	if err := confirmTx(xClient, assetID); err != nil {
		return err
	}

	expectedAVAXBalance = expectedAVAXBalance - aw.txFee

	logrus.Infof("Exporting AVAX")
	bech32CAddr := fmt.Sprintf("C%s", xAddr[1:])
	txID, err := xClient.ExportAVAX(user, nil, "", exportAVAXAmount, bech32CAddr)
	if err != nil {
		return fmt.Errorf("failed to export AVAX: %w", err)
	}
	if err := confirmTx(xClient, txID); err != nil {
		return err
	}
	expectedAVAXBalance -= (exportAVAXAmount + aw.txFee)

	logrus.Infof("Exporting asset")
	txID, err = xClient.Export(user, nil, "", exportAssetAmount, bech32CAddr, assetID.String())
	if err != nil {
		return fmt.Errorf("failed to export asset: %w", err)
	}
	if err := confirmTx(xClient, txID); err != nil {
		return err
	}
	expectedAVAXBalance -= aw.txFee
	expectedAssetBalance -= exportAssetAmount

	logrus.Infof("Importing to C Chain")
	txID, err = cClient.Import(user, cAddr, "X")
	if err != nil {
		return fmt.Errorf("failed to import from X-Chain: %w", err)
	}
	// TODO confirm the transaction instead of sleeping
	time.Sleep(2 * time.Second)
	if _, err = cClient.Import(user, cAddr, "X"); err == nil {
		return errors.New("importing a second time should have caused an error")
	}

	logrus.Infof("Verifying balances on X-Chain")
	// Confirm expected balance of AVAX and [assetID]
	balanceReply, err = xClient.GetBalance(xAddr, "AVAX")
	if err != nil {
		return fmt.Errorf("failed to get AVAX balance: %w", err)
	}
	foundAVAXBalance := uint64(balanceReply.Balance)
	if foundAVAXBalance != expectedAVAXBalance {
		return fmt.Errorf("expected AVAX balance of %d, but found %d", expectedAVAXBalance, foundAVAXBalance)
	}

	balanceReply, err = xClient.GetBalance(xAddr, assetID.String())
	if err != nil {
		return fmt.Errorf("failed to get asset balance: %w", err)
	}
	foundAssetBalance := uint64(balanceReply.Balance)
	if expectedAssetBalance != foundAssetBalance {
		return fmt.Errorf("expected asset balance of %d, but found %d", expectedAssetBalance, foundAssetBalance)
	}

	// Confirm Balances on C-Chain
	logrus.Infof("Verifying balances on C-Chain")
	ctx := context.Background()
	hexAddr := common.HexToAddress(cAddr)
	expectedCChainBalance := new(big.Int).Mul(big.NewInt(int64(x2cConversion)), big.NewInt(int64(exportAVAXAmount)))
	cBalance, err := cEthClient.BalanceAt(ctx, hexAddr, nil)
	if err != nil {
		return fmt.Errorf("failed to get AVAX balance on C-Chain: %w", err)
	}
	if cBalance.Cmp(expectedCChainBalance) != 0 {
		return fmt.Errorf("found unexpected balance %d, expected %d", cBalance, expectedCChainBalance)
	}
	cAssetBalance, err := cEthClient.AssetBalanceAt(ctx, hexAddr, assetID, nil)
	if err != nil {
		return fmt.Errorf("failed to get asset balance on C-Chain: %w", err)
	}
	bigExpectedAssetBalance := big.NewInt(int64(exportAssetAmount))
	if bigExpectedAssetBalance.Cmp(cAssetBalance) != 0 {
		return fmt.Errorf("found unexpected balance for asset: %d, expected %d", cAssetBalance, expectedAssetBalance)
	}

	logrus.Infof("Exporting back to X-Chain")
	txID, err = cClient.ExportAVAX(user, exportAVAXAmount, xAddr)
	if err != nil {
		return fmt.Errorf("failed to export AVAX to X-Chain: %w", err)
	}
	// TODO confirm tx
	time.Sleep(2 * time.Second)
	txID, err = cClient.Export(user, exportAssetAmount, xAddr, assetID.String())
	if err != nil {
		return fmt.Errorf("failed to export asset to X-Chain: %w", err)
	}
	// TODO confirm tx
	time.Sleep(2 * time.Second)

	// Confirm C-Chain balances are set back to 0
	logrus.Infof("Verifying C-Chain balances")
	zeroBalance := big.NewInt(0)
	cBalance, err = cEthClient.BalanceAt(ctx, hexAddr, nil)
	if err != nil {
		return fmt.Errorf("failed to get AVAX balance on C-Chain: %w", err)
	}
	if cBalance.Cmp(zeroBalance) != 0 {
		return fmt.Errorf("found unexpected balance %d, expected %d", cBalance, zeroBalance)
	}
	cAssetBalance, err = cEthClient.AssetBalanceAt(ctx, hexAddr, assetID, nil)
	if err != nil {
		return fmt.Errorf("failed to get asset balance on C-Chain: %w", err)
	}
	if cAssetBalance.Cmp(zeroBalance) != 0 {
		return fmt.Errorf("found unexpected balance for asset: %d, expected %d", cAssetBalance, zeroBalance)
	}

	// Import to X-Chain
	logrus.Infof("Importing funds back to X-Chain")
	txID, err = xClient.Import(user, xAddr, "C")
	if err != nil {
		return fmt.Errorf("failed to import from X -> C: %w", err)
	}
	if err := confirmTx(xClient, txID); err != nil {
		return err
	}

	expectedAVAXBalance -= aw.txFee
	expectedAssetBalance = assetAmount

	// Confirm Balances on X-Chain
	logrus.Infof("Verifying balances on the X-Chain")
	balanceReply, err = xClient.GetBalance(xAddr, "AVAX")
	if err != nil {
		return fmt.Errorf("failed to get AVAX balance: %w", err)
	}
	foundAVAXBalance = uint64(balanceReply.Balance)
	if foundAVAXBalance != expectedAVAXBalance {
		return fmt.Errorf("expected AVAX balance of %d, but found %d", expectedAVAXBalance, foundAVAXBalance)
	}

	balanceReply, err = xClient.GetBalance(xAddr, assetID.String())
	if err != nil {
		return fmt.Errorf("failed to get asset balance: %w", err)
	}
	foundAssetBalance = uint64(balanceReply.Balance)
	if expectedAssetBalance != foundAssetBalance {
		return fmt.Errorf("expected asset balance of %d, but found %d", expectedAssetBalance, foundAssetBalance)
	}

	return nil
}
