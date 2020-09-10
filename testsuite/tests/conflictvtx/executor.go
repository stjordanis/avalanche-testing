package conflictvtx

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche_client/apis"
	"github.com/ava-labs/avalanche-testing/testsuite/tester"
	"github.com/ava-labs/avalanche-go/snow/choices"
	"github.com/ava-labs/avalanche-go/utils/formatting"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type executor struct {
	virtuousClient  *apis.Client
	byzantineClient *apis.Client
}

// NewConflictingTxsVertexExecutor ...
func NewConflictingTxsVertexExecutor(virtuousClient, byzantineClient *apis.Client) tester.AvalancheTester {
	return &executor{
		virtuousClient:  virtuousClient,
		byzantineClient: byzantineClient,
	}
}

// ExecuteTest implements AvalancheTester interface
func (e *executor) ExecuteTest() error {
	byzantineXChainAPI := e.byzantineClient.XChainAPI()

	// TODO switch to test vectors or come up with method to reliably generate conflicting transactions
	// how to create a conflicting transaction???
	// test vector create asset tx and conflicting transactions
	// Note: these hardcoded transactions are based on everest-deployment with a txFee of 1,000,000 nAVAX
	cb58 := formatting.CB58{}
	if err := cb58.FromString("111118ErNmgStuDR9eVYXkdx1HbGsq4t7nSMq136wpmmsDU8ZAYv2W1A26yz2rnXeV1pzSSA7qNdKQqUnW3PNDgusduVzMNjSq183wbFJUGhYYvZEQ769aih5CUB2fHqu6vHc3cEXw3W6XotZhPEb82NZUMUNzpYSYpFrSi85eV46eDyYQGq49NFGGGZz7xhWsJyiiH3nWf3SwbGrkWmWRYdZUgwgKP2zqzvReu1yc9T41T2svGmDabA5KKKtzJWDP3JowGv9Dem53RPG8AQHkUY1Zn6s3atvgsz2a71dGggKdgBaJifE7Z9mknboaz2xkhL3VRzREEjcEquRBNg9RP9yQPv1ScXHFoF3PgrHW6y5RzPun59gGPrVg7Cfvh1n2rqpi73jB1nrLNyNFFXywtmeTv12p7Uz7qbEQ1ePP3rhJKciqmUcbqdGC9QajdhvKasQedCxPmFKVM9AnV789xUF5Z7QuXNraugc5cZG2HiuJc1XRyK6Yp4U6xmSie3umVGeuh"); err != nil {
		return stacktrace.Propagate(err, "Problem parsing create asset tx")
	}
	createAssetTx := cb58.Bytes
	if err := cb58.FromString("11111111BcUqA7u1jUPNAG82pPwFcGneFJihdTTr5ub5aUPwBoC7gtP3ZbSnJGaKGf6oJJMpd9LC21BjzzHE5Qtg6L5UMmkuHrmprvDPfqzpewRcoyuu5n8xKXv2kSmCTXUBXsURxZNSuZ2z3AQmg7xK3e91evtEVQNRUm8sSXpsxcBoESXKRygAAJVyUA2h54FHEi6gQtY4GsXjzDqXoxvktdpvUWynFwt4zWWjAiLmsq2A4aBUtzY5sTEs5KaYQ2Vf9sLdV9RLkFrDeGVdWBcqbYLyTY5eYp6a7J8HwurXACF5iEzwvxBMByhUqmVSMCbWcuiTWTWTU98SZTzH5PVHufGKEJCk269pZMRe6M1oUNhqsGMgQ6A5RSKzrT5V5dkAikvjt4K9Fs72tGbk4Lxmmy7dSDToRXGskHqccm4rQQjtdCbYVZQRMu6dqgrQkGa3mNkFPqHZDvgryPL288NJEsfmjR2SWTR8Eco5nRJ4FzZmViCXFVsPeLuLhsKSAWYTzaWNWEEsrPvz2MymW37gmK9pufsXVqnFZehfHS68VvJDW1LNLjV2dqDPXmH8Aep3PorqG9SqYCA5FPjts8qk1YkAFjjRPKSNehcbocNMBvnsPwHH4dJjDLfwYHet2ncs7zZSLqhqnj1kv37JgNSumqB9UBbAqavNycdcRjAUDSRSdYm4wuNd1PZx176Y1NkMeRGWphn7d15cYSLdcs7gUQ148iXxxPzu"); err != nil {
		return stacktrace.Propagate(err, "Problem parsing first send tx")
	}
	conflictingTx1 := cb58.Bytes

	if err := cb58.FromString("11111111BcUqA7u1jUPNAG82pPwFcGneFJihdTTr5ub5aUPwBoC7gtP3ZbSnJGaKGf6oJJMpd9LC21BjzzHE5Qtg6L5UMmkuHrmprvDPfqzpewRcoyuu5n8xKXv2kSmCTXUBXsURxZNSuZ2z3AQmg7xK3e91evtEVQNRUm8sSXpsxcBoESXKRygAAJVyUA2h54FHEi6gQtY4GsXjzDqXoxvktdpvUWynFwt4zWWjAiLmsq2A4aBUtzY5sTFMpLDeMj74Bdp99ueWJfYP5iooGm7osougGL3SogSfhcsTxQiLaNsPa2yUB9JjoFkZwMgyA75frQRKbMgcLoyT7beZwkZF7gTG3VCaEWCeZS9L9M96rdMi1FDTSm95AN4miUBzynVXsgJcUKhaz3qGkr72dkrr2gxi5zPpBekW5cZMyB7MUAggyMn9RyEgBBoBS8U5PwZPgNQnfMhBrmEUQjJ4L1kKCbT2t1cQPcoJNEzSL56eoJxtWvG1rnhnrkmGSkQVy5J3bXoyL5twwwz8zDpEuvNcQ9dacNP6NJcMS8yMpf55ARrHH6GNvQ6ekLrftBQCfM5wKbs4TtUEgXEvUSbkAfsecbbmY2iHUTvEdj1bPba2fXJ6XoApShjKgAyoMmYdzgSVtdq5ffQLmK2zmGrf3GVdHauCfzzj87S2DTf2DtRHbEyFrDYpbZ2VaE1FdhyVoH5KjfLa2XbePhKZyeBFGfzjq9YmvDCmpi7k"); err != nil {
		return stacktrace.Propagate(err, "Problem parsing second send tx")
	}
	conflictingTx2 := cb58.Bytes

	logrus.Infof("Issuing conflicting transactions to a byzantine node...")
	nonConflictID, err := byzantineXChainAPI.IssueTx(createAssetTx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue first transaction to byzantine node.")
	}
	conflictID1, err := byzantineXChainAPI.IssueTx(conflictingTx1)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue second transaction to byzantine node.")
	}
	conflictID2, err := byzantineXChainAPI.IssueTx(conflictingTx2)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue third transaction to byzantine node.")
	}

	logrus.Infof("Issued three transactions to Byzantine Node with IDs: %s, %s, %s", nonConflictID, conflictID1, conflictID2)

	// Confirm the byzantine node Accepted the transactions
	// Note: The byzantine behavior is to batch the pending transactions into a vertex as soon as it detects a conflict.
	// It should try to accept each transaction before PushQuery-ing the vertex to other nodes to signal to this test
	// controller that the vertex was successfully issued
	status, err := byzantineXChainAPI.GetTxStatus(nonConflictID)
	if err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", nonConflictID))
	}
	if status != choices.Accepted {
		return stacktrace.Propagate(err, fmt.Sprintf("Transaction: %s was not accepted, status: %s", nonConflictID, status))
	}

	logrus.Infof("Status of non-conflict transactions on byzantine node is: %s", status)

	conflictStatus1, err := byzantineXChainAPI.GetTxStatus(conflictID1)
	if err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", conflictID1))
	}

	logrus.Infof("Status of conflict tx1: %s on byzantine node is: %s", conflictID1, conflictStatus1)

	conflictStatus2, err := byzantineXChainAPI.GetTxStatus(conflictID2)
	if err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", conflictID2))
	}

	logrus.Infof("Status of conflict tx2: %s on byzantine node is: %s", conflictID2, conflictStatus2)

	// Byzantine node should try to accept both conflicting transactions, but will fail to accept one due to the missing UTXO
	// after the other consumes it.
	if conflictStatus1 != choices.Accepted && conflictStatus2 != choices.Accepted {
		return fmt.Errorf("Byzantine node did not accept either of the conflicting transactions, status1: %s. status2: %s", conflictStatus1, conflictStatus2)
	}

	// The issued vertex should be dropped completely, so the virtuous nodes should drop the vertex
	// and never issue the transactions into consensus.
	// Note: since the transactions will be parsed in the process, we expect the status to be "Processing" not "Rejected" or "Unknown"
	// We issue a vertex from the virtuous node to see if it builds on invalid vertex
	// This is meant to remove the need to wait an arbitrary amount of time to see if the vertex gets accepted
	// and instead confirm the valid transaction as a measure of the time to finality before checking if
	// the transactions that should have been dropped were in fact dropped successfully.
	// TODO move to test vector
	virtuousXChainAPI := e.virtuousClient.XChainAPI()
	if err := cb58.FromString("111118ErNmgStuDR9eVYXkdx1HbGsq4t7nSMq136wpmmsDU8ZAYv2W1A26yz2rnXeV1pzSSA7qNdKQqUnW3PNDgusduVzMNjSq183wbFJUGhYYvZEQ769aih5CUB2fHqu6vHc3cEXw3W6XotZhPEb82NZUMUNzpYSYpFrSi85eV46eDyYQGq49NFGGGZz7xhWsJyiiH3nWf3SwbGrkWmWRYdZUgwgKP2zqzvReu1yc9T41T2svGmDabA5KKKtzJWDP3JowGv9Dem53RPG8AQHkUY1Zn6s3atvgsz2a71dGggKdgBaJifE7Z9mknboaz2xkhL3VRzREEjcEquRBNg9RP9yQPv1ScXHFoF2n7XcphjFd3ER7y1uRjDffbe5AQh4nFKZEQEfcgYcshAcXAokuGBiGnBm6JapPd813F4hdXjF6zQJScg99KE9s5nR5w7mGBbqvoenn1MTQCfj9CRxsUqqDAaLL38khHhWQN46osZxLV67qvJMfQ3FWjoY8whnsDZfbV"); err != nil {
		return stacktrace.Propagate(err, "Failed to parse virtuous create asset tx")
	}
	virtuousCreateAssetTx := cb58.Bytes
	if err := cb58.FromString("11111111BcUqA7u1jUPNAG82pPwFcGneFJihdTTr5ub5aUPwBoC7gtP3ZbSnJGaKGf6oJJMpd9LC21BjzzHE5Qtg6L5UMmkuHrmprvDPfqzpewRcoyuu5n8xKXv2kSmCTXUBXsURxZNSuZ2z3AQmg7xK3e91evtEVQNRUm8sShjK6pMPDpQsPaNDPTbTMZwwryRmy9KiYBjnjyHM8kFwbCkP7Lr9AyVHU5p3YKfAa3yFuEPfGRC277Lmrt6fMUWf5LTB3ivdXiXUF1zU5QbRDSx7WUVzewJhhNW6yCKcAbLpL3B3yvVsjGTspLkLbXtnfYEY4yiGqKt8vv9fA62CQoA9KfJJRW316Ymv8qgERt4bwnSot8tfmAUgmFQQs5gQwfXLtSXbiyCAeATUysheLN1SGGmqRwaRLYU5tYxGNMdSnNZKGhQpLyxopRArLE5jhSWke7Aihx6w4MQw1ff3yFRiCStQC1onCcCQSqRkUob4UvQFoankxhKmrVqPqTqb4DH6MvBimV2w7VqXaT8gzAQWGrhvnEcRJFSr9ZcY62HL2zVisweSxHurcpeGBfHLAhD7rWUrVkTaNrwzu3KZhrFqH3ae1ZLPZ2wu6euUyV4QWWvXEA7RSfp2uDU5cnnddsVcop1dfNNPmq5gd2wTziDEvFZPSHb6kG4Pn9vABsfheeE3iTRS1LRAfabFurxifAPzCFMVtH9XUxWXvXQsjXztLikUW9eLtrGV"); err != nil {
		return stacktrace.Propagate(err, "Failed to parse virtuous spend tx")
	}
	virtuousSpendTx := cb58.Bytes

	// Ignore the TxID of this because it should be accepted immediately after entering consensus
	_, err = virtuousXChainAPI.IssueTx(virtuousCreateAssetTx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue virtuous create asset transaction after issuing illegal vertex from byzantine node.")
	}
	virtuousSpendTxID, err := virtuousXChainAPI.IssueTx(virtuousSpendTx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue virtuous transaction spending created asset after issuing byzantine vertex")
	}

	for {
		status, err := virtuousXChainAPI.GetTxStatus(virtuousSpendTxID)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get virtuous transactions status from virtuous node")
		}
		if status == choices.Accepted {
			logrus.Infof("Accepted virtuous transaction with ID: %s", virtuousSpendTxID)
			break
		} else {
			logrus.Infof("Waiting for transaction with ID: %s to be accepted", virtuousSpendTxID)
			time.Sleep(2 * time.Second)
		}
	}

	// Once the virtuous transaction was accepted, check to see if the non-conflicting transaction
	// in an illegal vertex was accepted
	status, err = virtuousXChainAPI.GetTxStatus(nonConflictID)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get transaction status for non-conflicting transaction.")
	}
	logrus.Infof("Status of CreateAssetTx: %s is %s", nonConflictID, status)
	// If the transaction was Accepted, the test should fail because virtuous nodes should not issue the vertex and
	// the underlying transactions into consensus
	if status == choices.Accepted {
		return stacktrace.Propagate(err, fmt.Sprintf("Expected status of non-conflicting transaction issued in bad vertex to be Processing, but found %s", status))
	}
	return nil
}
