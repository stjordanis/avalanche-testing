package bombard

import (
	"fmt"

	"github.com/ava-labs/avalanche-testing/utils/constants"
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/codec/hierarchycodec"
	"github.com/ava-labs/avalanchego/codec/linearcodec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/nftfx"
	"github.com/ava-labs/avalanchego/vms/propertyfx"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

const (
	// Codec version used before AvalancheGo 1.1.0
	pre110CodecVersion = uint16(0)

	// Current codec version
	currentCodecVersion = uint16(1)
)

func createXChainCodec() (codec.Manager, error) {
	codecManager := codec.NewDefaultManager()

	pre110Codec := linearcodec.NewDefault()
	errs := wrappers.Errs{}
	errs.Add(
		pre110Codec.RegisterType(&avm.BaseTx{}),
		pre110Codec.RegisterType(&avm.CreateAssetTx{}),
		pre110Codec.RegisterType(&avm.OperationTx{}),
		pre110Codec.RegisterType(&avm.ImportTx{}),
		pre110Codec.RegisterType(&avm.ExportTx{}),
		pre110Codec.RegisterType(&secp256k1fx.TransferInput{}),
		pre110Codec.RegisterType(&secp256k1fx.MintOutput{}),
		pre110Codec.RegisterType(&secp256k1fx.TransferOutput{}),
		pre110Codec.RegisterType(&secp256k1fx.MintOperation{}),
		pre110Codec.RegisterType(&secp256k1fx.Credential{}),
		pre110Codec.RegisterType(&nftfx.MintOutput{}),
		pre110Codec.RegisterType(&nftfx.TransferOutput{}),
		pre110Codec.RegisterType(&nftfx.MintOperation{}),
		pre110Codec.RegisterType(&nftfx.TransferOperation{}),
		pre110Codec.RegisterType(&nftfx.Credential{}),
		pre110Codec.RegisterType(&propertyfx.MintOutput{}),
		pre110Codec.RegisterType(&propertyfx.OwnedOutput{}),
		pre110Codec.RegisterType(&propertyfx.MintOperation{}),
		pre110Codec.RegisterType(&propertyfx.BurnOperation{}),
		pre110Codec.RegisterType(&propertyfx.Credential{}),
		codecManager.RegisterCodec(pre110CodecVersion, pre110Codec),
	)
	if errs.Errored() {
		return nil, errs.Err
	}

	currentCodec := hierarchycodec.NewDefault()
	errs.Add(
		currentCodec.RegisterType(&avm.BaseTx{}),
		currentCodec.RegisterType(&avm.CreateAssetTx{}),
		currentCodec.RegisterType(&avm.OperationTx{}),
		currentCodec.RegisterType(&avm.ImportTx{}),
		currentCodec.RegisterType(&avm.ExportTx{}),
	)
	currentCodec.NextGroup()
	errs.Add(
		currentCodec.RegisterType(&secp256k1fx.TransferInput{}),
		currentCodec.RegisterType(&secp256k1fx.MintOutput{}),
		currentCodec.RegisterType(&secp256k1fx.TransferOutput{}),
		currentCodec.RegisterType(&secp256k1fx.MintOperation{}),
		currentCodec.RegisterType(&secp256k1fx.Credential{}),
		currentCodec.RegisterType(&secp256k1fx.ManagedAssetStatusOutput{}),
		currentCodec.RegisterType(&secp256k1fx.UpdateManagedAssetOperation{}),
	)
	currentCodec.NextGroup()
	errs.Add(
		currentCodec.RegisterType(&nftfx.MintOutput{}),
		currentCodec.RegisterType(&nftfx.TransferOutput{}),
		currentCodec.RegisterType(&nftfx.MintOperation{}),
		currentCodec.RegisterType(&nftfx.TransferOperation{}),
		currentCodec.RegisterType(&nftfx.Credential{}),
	)
	currentCodec.NextGroup()
	errs.Add(
		currentCodec.RegisterType(&propertyfx.MintOutput{}),
		currentCodec.RegisterType(&propertyfx.OwnedOutput{}),
		currentCodec.RegisterType(&propertyfx.MintOperation{}),
		currentCodec.RegisterType(&propertyfx.BurnOperation{}),
		currentCodec.RegisterType(&propertyfx.Credential{}),
		codecManager.RegisterCodec(currentCodecVersion, currentCodec),
	)
	return codecManager, errs.Err
}

// CreateSingleUTXOTx returns a transaction spending an individual utxo owned by [privateKey]
func CreateSingleUTXOTx(utxo *avax.UTXO, inputAmount, outputAmount uint64, address ids.ShortID, privateKey *crypto.PrivateKeySECP256K1R, codec codec.Manager) (*avm.Tx, error) {
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
		NetworkID:    constants.NetworkID,
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
		inputAmount -= txFee
		outputAmount -= txFee
	}

	return txBytes, txIDs, nil
}
