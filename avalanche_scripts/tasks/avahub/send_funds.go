package avahub

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	avmClient "github.com/ava-labs/avalanche-testing/avalanche_client/apis/avm"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/keystore"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/snow/choices"
	cjson "github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/sirupsen/logrus"
)

var (
	uri            = "http://127.0.0.1:9650"
	requestTimeout = 5 * time.Second
	user           = api.UserPass{Username: "aaronb2", Password: "eur1hiybweiyer"}
	privateKey     = "PrivateKey-DVm6PGG2qnmuozQZZPwqMQrFh2a2EXEHMdE7QrP2VX6n23Z64"
	// privateKey = "PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
)

// CreateOutputsFromCSV ...
func CreateOutputsFromCSV(fileName string) ([]avm.SendOutput, uint64, error) {
	totalAmount := uint64(0)

	file, err := os.Open(fileName)
	if err != nil {
		return nil, 0, err
	}
	reader := csv.NewReader(file)

	outputs := make([]avm.SendOutput, 0)
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

		address := record[1]

		// address := record[1]
		nAVAX, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			return nil, 0, err
		}
		outputs = append(outputs, avm.SendOutput{
			To:      address,
			Amount:  cjson.Uint64(nAVAX),
			AssetID: "AVAX",
		})
	}

	return outputs, totalAmount, nil
}

// SendFunds ...
func SendFunds(fileName string) {
	outputs, totalAmount, err := CreateOutputsFromCSV(fileName)
	if err != nil {
		logrus.Errorf("Failed to create outputs from CSV: %s", err)
		return
	}
	logrus.Infof("Created outputs totaling: %d", totalAmount)

	keystoreClient := keystore.NewClient(uri, requestTimeout)
	_, err = keystoreClient.CreateUser(user)
	if err != nil {
		logrus.Warnf("Failed to create user: %s", err)
	}
	client := avmClient.NewClient(uri, "X", requestTimeout)

	addr, err := client.ImportKey(user, privateKey)
	if err != nil {
		logrus.Errorf("Failed to import key: %s", err)
		return
	}
	logrus.Infof("Address: %s", addr)
	txID, err := client.SendMultiple(user, nil, "", outputs, "Avalanche Hub")
	if err != nil {
		logrus.Errorf("Failed to send transaction: %s", err)
		return
	}
	logrus.Infof("Sent transaction with ID: %s", txID)

	for {
		time.Sleep(time.Second)
		status, err := client.GetTxStatus(txID)
		if err != nil {
			logrus.Errorf("Failed to get tx status: %s", err)
			return
		}
		if status == choices.Accepted {
			break
		}
		logrus.Infof("Status is: %s", status)
	}

	for _, output := range outputs {
		balanceReply, err := client.GetBalance(output.To, "AVAX")
		if err != nil {
			logrus.Errorf("Failed to get balance for %s: %s", output.To, err)
			return
		}

		// if uint64(balanceReply.Balance) != uint64(output.Amount) {
		// 	logrus.Errorf("%s", fmt.Sprintf("incorrect balance: %d, expected: %d. Address: %s", balanceReply.Balance, output.Amount, output.To))
		// }
		if uint64(balanceReply.Balance) < uint64(output.Amount) {
			logrus.Errorf("Should have sent at least: %d to %s, but found balance of %d", uint64(output.Amount), output.To, uint64(balanceReply.Balance))
			return
		}
		logrus.Infof("Sent %d to address: %s. Balance was: %d", uint64(output.Amount), output.To, uint64(balanceReply.Balance))
	}
}
