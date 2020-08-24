package admin_rpc

import (
	"fmt"

	"github.com/ava-labs/avalanche-testing/gecko_client/utils/constants"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/utils/codec"
	"github.com/ava-labs/gecko/utils/crypto"
	"github.com/ava-labs/gecko/utils/wrappers"
	"github.com/ava-labs/gecko/vms/avm"
	"github.com/ava-labs/gecko/vms/components/avax"
	"github.com/ava-labs/gecko/vms/propertyfx"
	"github.com/ava-labs/gecko/vms/secp256k1fx"
)

const (
	networkID uint32 = 12345 // TODO move to constants package
)

func createXChainCodec() (codec.Codec, error) {
	c := codec.NewDefault()
	errs := wrappers.Errs{}
	errs.Add(
		c.RegisterType(&avm.BaseTx{}),
		c.RegisterType(&avm.CreateAssetTx{}),
		c.RegisterType(&avm.OperationTx{}),
		c.RegisterType(&avm.ImportTx{}),
		c.RegisterType(&avm.ExportTx{}),

		c.RegisterType(&secp256k1fx.TransferInput{}),
		c.RegisterType(&secp256k1fx.MintOutput{}),
		c.RegisterType(&secp256k1fx.TransferOutput{}),
		c.RegisterType(&secp256k1fx.MintOperation{}),
		c.RegisterType(&secp256k1fx.Credential{}),

		c.RegisterType(&propertyfx.MintOutput{}),
		c.RegisterType(&propertyfx.OwnedOutput{}),
		c.RegisterType(&propertyfx.MintOperation{}),
		c.RegisterType(&propertyfx.BurnOperation{}),
		c.RegisterType(&propertyfx.Credential{}),
	)

	return c, errs.Err
}

// CreateSingleUTXOTx returns a transaction spending an individual utxo owned by [privateKey]
func CreateSingleUTXOTx(utxo *avax.UTXO, inputAmount, outputAmount uint64, address ids.ShortID, privateKey *crypto.PrivateKeySECP256K1R, codec codec.Codec) (*avm.Tx, error) {
	keys := [][]*crypto.PrivateKeySECP256K1R{{privateKey}}
	outs := []*avax.TransferableOutput{&avax.TransferableOutput{
		Asset: avax.Asset{ID: constants.AvaxAssetID},
		Out: &secp256k1fx.TransferOutput{
			Amt: outputAmount,
			OutputOwners: secp256k1fx.OutputOwners{
				Locktime:  0,
				Threshold: 1,
				Addrs:     []ids.ShortID{address},
			},
		},
	}}

	transferableIn := interface{}(&secp256k1fx.TransferInput{
		Amt: inputAmount,
		Input: secp256k1fx.Input{
			SigIndices: []uint32{0},
		},
	})

	ins := []*avax.TransferableInput{&avax.TransferableInput{
		UTXOID: utxo.UTXOID,
		Asset:  avax.Asset{ID: constants.AvaxAssetID},
		In:     transferableIn.(avax.TransferableIn),
	}}

	tx := &avm.Tx{UnsignedTx: &avm.BaseTx{BaseTx: avax.BaseTx{
		NetworkID:    networkID,
		BlockchainID: constants.XChainID,
		Outs:         outs,
		Ins:          ins,
	}}}

	if err := tx.SignSECP256K1Fx(codec, keys); err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateConsecutiveTransactions returns a string of [numTxs] sending [utxo] back and forth
// assumes that [privateKey] is the sole owner of [utxo]
func CreateConsecutiveTransactions(utxo *avax.UTXO, numTxs, amount, txFee uint64, privateKey *crypto.PrivateKeySECP256K1R) ([][]byte, []ids.ID, error) {
	if numTxs*txFee > amount {
		return nil, nil, fmt.Errorf("Insufficient starting funds to send %v transactions with a txFee of %v", numTxs, txFee)
	}
	codec, err := createXChainCodec()
	if err != nil {
		return nil, nil, err
	}

	address := privateKey.PublicKey().Address()
	txBytes := make([][]byte, numTxs)
	txIDs := make([]ids.ID, numTxs)

	inputAmount := amount
	outputAmount := amount - txFee
	for i := uint64(0); i < numTxs; i++ {
		tx, err := CreateSingleUTXOTx(utxo, inputAmount, outputAmount, address, privateKey, codec)
		if err != nil {
			return nil, nil, err
		}
		txBytes[i] = tx.Bytes()
		txIDs[i] = tx.ID()
		utxo = tx.UTXOs()[0]
		inputAmount = inputAmount - txFee
		outputAmount = outputAmount - txFee
	}

	return txBytes, txIDs, nil
}
