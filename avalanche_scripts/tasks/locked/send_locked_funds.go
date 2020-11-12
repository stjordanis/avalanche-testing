package locked

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/avm"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/keystore"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/platform"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/math"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/sirupsen/logrus"
)

const (
	networkID = uint32(1)
	txFee     = uint64(1000000)
	pChainID  = "P"
	bech32hrp = "avax"
	// fundedPrivateKey = "PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
	// fundedPrivateKey = "PrivateKey-23tYfNBVmsvMeEekJhzwtduuzPbqBeJWacj9obeXhiFQVxcA5v"
	// fundedPrivateKey = "PrivateKey-DVm6PGG2qnmuozQZZPwqMQrFh2a2EXEHMdE7QrP2VX6n23Z64"
	fundedPrivateKey = "PrivateKey-PFCgJA9NSDqBWj13DihoyQKLabYVnX1BgDiDqa3izB2QZaWHP"
	// fundedPrivateKey = "PrivateKey-tyvQ6wYWnx3kbXJUeSbs9nqZUwGrdQWsyueNHPhLMVtfr3STe"
	uri            = "http://127.0.0.1:9650"
	requestTimeout = 5 * time.Second
)

var (
	platformChainID                              = ids.Empty
	xChainID                                     = ids.Empty
	avaxAssetID                                  = ids.Empty
	privateKey      *crypto.PrivateKeySECP256K1R = nil
	user                                         = api.UserPass{Username: "sendlockedfunds", Password: "TheButtFumbleWasStaged6"}
)

func init() {
	privateKey, _ = ConvertPrivateKey(fundedPrivateKey)
	xChainID, _ = ids.FromString("2oYMBNV4eNHyqk2fjjV5nVQLDbtmNJzq5s3qs3Lo6ftnC6FByM")
	avaxAssetID, _ = ids.FromString("FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z")
}

// ConvertPrivateKey ...
func ConvertPrivateKey(privateKey string) (*crypto.PrivateKeySECP256K1R, error) {
	if !strings.HasPrefix(privateKey, constants.SecretKeyPrefix) {
		return nil, fmt.Errorf("private key missing %s prefix", constants.SecretKeyPrefix)
	}

	trimmedPrivateKey := strings.TrimPrefix(privateKey, constants.SecretKeyPrefix)
	formattedPrivateKey := formatting.CB58{}
	if err := formattedPrivateKey.FromString(trimmedPrivateKey); err != nil {
		return nil, fmt.Errorf("problem parsing private key: %w", err)
	}

	factory := crypto.FactorySECP256K1R{}
	skIntf, err := factory.ToPrivateKey(formattedPrivateKey.Bytes)
	if err != nil {
		return nil, fmt.Errorf("problem parsing private key: %w", err)
	}
	sk := skIntf.(*crypto.PrivateKeySECP256K1R)
	return sk, nil
}

// ExportRequiredAVAX ...
func ExportRequiredAVAX(avm *avm.Client, user api.UserPass, pAddress string, requiredAmount uint64) error {
	txID, err := avm.ExportAVAX(user, requiredAmount, pAddress, nil, "")
	if err != nil {
		return fmt.Errorf("Failed to ExportAVAX: %w", err)
	}
	logrus.Infof("Exported AVAX to %s, TxID: %s", pAddress, txID)

	for {
		time.Sleep(time.Second)
		status, err := avm.GetTxStatus(txID)
		if err != nil {
			return fmt.Errorf("Failed to get tx status for %s: %w", txID, err)
		}
		if status == choices.Accepted {
			break
		}
		logrus.Infof("Tx status: %s", status)
	}

	return nil
}

// CreateStakeableLockOutput ...
func CreateStakeableLockOutput(address ids.ShortID, amount uint64, locktime uint64) *avax.TransferableOutput {
	if locktime == 0 {
		return &avax.TransferableOutput{
			Asset: avax.Asset{ID: avaxAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: amount,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{address},
				},
			},
		}
	}

	return &avax.TransferableOutput{
		Asset: avax.Asset{ID: avaxAssetID},
		Out: &platformvm.StakeableLockOut{
			Locktime: locktime,
			TransferableOut: &secp256k1fx.TransferOutput{
				Amt: amount,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{address},
				},
			},
		},
	}
}

// CreateLockedOutputsFromCSV ...
func CreateLockedOutputsFromCSV(fileName string) ([]*avax.TransferableOutput, uint64, error) {
	totalAmount := uint64(0)

	file, err := os.Open(fileName)
	if err != nil {
		return nil, 0, err
	}
	reader := csv.NewReader(file)
	// Assumes there is a header to the csv file
	// _, err = reader.Read()
	// if err != nil {
	// 	return nil, 0, err
	// }

	outputs := make([]*avax.TransferableOutput, 0, 10) // Adjust estimate here to optimize memory allocation
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to read due to %s", err)
		}
		if len(record) != 3 {
			return nil, 0, fmt.Errorf("Expected to find record of length 3, but found length: %d", len(record))
		}

		// TODO parse the bech32 address
		// address := record[0]
		_, hrp, b, err := formatting.ParseAddress(record[0])
		if err != nil {
			return nil, 0, err
		}
		// if chainID != pChainID {
		// 	return nil, 0, fmt.Errorf("incorrect chainID: %s, expected: %s", chainID, pChainID)
		// }
		if hrp != bech32hrp {
			return nil, 0, fmt.Errorf("incorrect hrp in address: %s, expected: %s", hrp, bech32hrp)
		}
		address, err := ids.ToShortID(b)
		if err != nil {
			return nil, 0, err
		}

		amount, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse amount due to: %w", err)
		}

		locktime, err := strconv.ParseUint(record[1], 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse locktime due to: %w", err)
		}

		outputs = append(outputs, CreateStakeableLockOutput(address, amount, locktime))
		totalAmount += amount
	}

	return outputs, totalAmount, nil
}

// BatchFundOutputs ...
func BatchFundOutputs(avm *avm.Client, platformClient *platform.Client, user api.UserPass, pAddress string, outputs []*avax.TransferableOutput) error {
	totalAmount := uint64(0)
	for _, output := range outputs {
		totalAmount += output.Out.Amount()
	}

	totalAmount += txFee
	// wrappedOutputs is now the outputs that we want to fund
	if err := ExportRequiredAVAX(avm, user, pAddress, totalAmount); err != nil {
		return fmt.Errorf("failed to export the required AVAX: %w", err)
	}
	logrus.Infof("Exported %d AVAX to P Chain", totalAmount)

	allUTXOBytes, err := platformClient.GetUTXOs([]string{pAddress}, "X")
	if err != nil {
		return fmt.Errorf("Failed to retrieve UTXOs for %s due to %w", pAddress, err)
	}

	utxos := make([]*avax.UTXO, len(allUTXOBytes))
	for i, utxoBytes := range allUTXOBytes {
		utxo := &avax.UTXO{}
		if err := platformvm.Codec.Unmarshal(utxoBytes, utxo); err != nil {
			return fmt.Errorf("failed to unmarshal UTXO: %w", err)
		}
		utxos[i] = utxo
	}
	logrus.Infof("Fetched %d UTXOs", len(utxos))

	// We need enough inputs to fund totalAmount
	kc := secp256k1fx.NewKeychain()
	kc.Add(privateKey)

	importedInputs := []*avax.TransferableInput{}
	signers := [][]*crypto.PrivateKeySECP256K1R{}

	importedAmount := uint64(0)
	now := uint64(time.Now().Unix())
	for _, utxo := range utxos {
		if !utxo.AssetID().Equals(avaxAssetID) {
			continue
		}
		inputIntf, utxoSigners, err := kc.Spend(utxo.Out, now)
		if err != nil {
			continue
		}
		input, ok := inputIntf.(avax.TransferableIn)
		if !ok {
			continue
		}
		importedAmount, err = math.Add64(importedAmount, input.Amount())
		if err != nil {
			return err
		}
		importedInputs = append(importedInputs, &avax.TransferableInput{
			UTXOID: utxo.UTXOID,
			Asset:  utxo.Asset,
			In:     input,
		})
		signers = append(signers, utxoSigners)
	}
	avax.SortTransferableInputsWithSigners(importedInputs, signers)

	if importedAmount > totalAmount {
		outputs = append(outputs, &avax.TransferableOutput{
			Asset: avax.Asset{ID: avaxAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: importedAmount - totalAmount,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{privateKey.PublicKey().Address()},
				},
			},
		})
	} else if importedAmount < totalAmount {
		return fmt.Errorf("Only found imported amount of %d, needed %d", importedAmount, totalAmount)
	}

	avax.SortTransferableOutputs(outputs, platformvm.Codec)

	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedImportTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    networkID,
			BlockchainID: platformChainID,
			Outs:         outputs,
			Ins:          []*avax.TransferableInput{},
		}},
		SourceChain:    xChainID,
		ImportedInputs: importedInputs,
	}}
	if err := tx.Sign(platformvm.Codec, signers); err != nil {
		return err
	}

	txBytes, err := platformvm.Codec.Marshal(tx)
	if err != nil {
		return err
	}
	logrus.Infof("Created transaction")

	txID, err := platformClient.IssueTx(txBytes)
	if err != nil {
		return err
	}

	logrus.Infof("Issued ImportTx with ID: %s", txID)

	for {
		time.Sleep(time.Second)
		status, err := platformClient.GetTxStatus(txID)
		if err != nil {
			return fmt.Errorf("Failed to get tx status for %s: %w", txID, err)
		}
		if status == platformvm.Committed {
			logrus.Infof("Transaction was committed")
			break
		}
		logrus.Infof("Tx status: %s", status)
	}

	return nil
}

// SendLockedFunds ...
func SendLockedFunds(fileName string) {
	keystore := keystore.NewClient(uri, requestTimeout)
	if _, err := keystore.CreateUser(user); err != nil {
		logrus.Warnf("didn't create new user")
	}
	outputs, totalAmount, err := CreateLockedOutputsFromCSV(fileName)
	if err != nil {
		logrus.Errorf("Failed to create outputs from CSV: %s", err)
		return
	}

	logrus.Infof("Created %d outputs with a total amount of %d", len(outputs), totalAmount)
	avm := avm.NewClient(uri, "X", requestTimeout)
	xAddress, err := avm.ImportKey(user, fundedPrivateKey)
	if err != nil {
		logrus.Errorf("Failed to import private key: %s", err)
		return
	}
	logrus.Infof("Imported key with address: %s", xAddress)
	platformClient := platform.NewClient(uri, requestTimeout)
	pAddress, err := platformClient.ImportKey(user, fundedPrivateKey)
	if err != nil {
		logrus.Errorf("failed to import platform key: %w", err)
		return
	}

	logrus.Infof("Imported key to paddress: %s", pAddress)

	if err := BatchFundOutputs(avm, platformClient, user, pAddress, outputs); err != nil {
		logrus.Errorf("Failed to batch fund outputs: %s", err)
	}
}
