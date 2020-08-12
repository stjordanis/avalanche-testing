package conflicting_txs_vertex_test

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type executor struct {
	virtuousClient  *apis.Client
	byzantineClient *apis.Client
}

func NewConflictingTxsVertexExecutor(virtuousClient, byzantineClient *apis.Client) ava_testsuite.AvalancheTester {
	return &executor{
		virtuousClient:  virtuousClient,
		byzantineClient: byzantineClient,
	}
}

func (e *executor) ExecuteTest() error {
	byzantineXChainAPI := e.byzantineClient.XChainAPI()
	// TODO switch to test vectors or come up with method to reliably generate conflicting transactions
	// how to create a conflicting transaction???
	// test vector create asset tx and conflicting transactions
	createAssetTx := []byte{1}
	conflictingTx1 := []byte{2}
	conflictingTx2 := []byte{3}

	nonConflictId, err := byzantineXChainAPI.IssueTx(createAssetTx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue first transaction to byzantine node.")
	}
	conflictId1, err := byzantineXChainAPI.IssueTx(conflictingTx1)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue second transaction to byzantine node.")
	}
	conflictId2, err := byzantineXChainAPI.IssueTx(conflictingTx2)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue third transaction to byzantine node.")
	}

	logrus.Infof("Issued three transactions to Byzantine Node with IDs: %s, %s, %s", nonConflictId, conflictId1, conflictId2)

	// Confirm the byzantine node Accepted the transactions
	// Note: The byzantine behavior is to batch the pending transactions into a vertex as soon as it detects a conflict.
	// It should try to accept each transaction before PushQuery-ing the vertex to other nodes to signal to this test
	// controller that the vertex was successfully issued
	status, err := byzantineXChainAPI.GetTxStatus(nonConflictId)
	if err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", nonConflictId))
	}
	if status != choices.Accepted {
		return stacktrace.Propagate(err, fmt.Sprintf("Transaction: %s was not accepted, status: %s", nonConflictId, status))
	}

	logrus.Infof("Status of non-conflict transactions on byzantine node is: %s", status)

	conflictStatus1, err := byzantineXChainAPI.GetTxStatus(conflictId1)
	if err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", conflictId1))
	}

	logrus.Infof("Status of conflict tx1: %s on byzantine node is: %s", conflictId1, conflictStatus1)

	conflictStatus2, err := byzantineXChainAPI.GetTxStatus(conflictId2)
	if err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", conflictId2))
	}

	logrus.Infof("Status of conflict tx2: %s on byzantine node is: %s", conflictId2, conflictStatus2)

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
	virtuousCreateAssetTx := []byte{4}
	virtuousSpendTx := []byte{5}

	// Ignore the TxID of this because it should be accepted immediately after entering consensus
	_, err = virtuousXChainAPI.IssueTx(virtuousCreateAssetTx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue virtuous create asset transaction after issuing illegal vertex from byzantine node.")
	}
	virtuousSpendTxId, err := virtuousXChainAPI.IssueTx(virtuousSpendTx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue virtuous transaction spending created asset after issuing byzantine vertex")
	}

	for {
		status, err := virtuousXChainAPI.GetTxStatus(virtuousSpendTxId)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get virtuous transactions status from virtuous node")
		}
		if status == choices.Accepted {
			logrus.Infof("Accepted virtuous transaction with ID: %s", virtuousSpendTxId)
			break
		} else {
			logrus.Infof("Waiting for transaction with ID: %s to be accepted", virtuousSpendTxId)
			time.Sleep(2 * time.Second)
		}
	}

	// Once the virtuous transaction was accepted, check to see if the non-conflicting transaction
	// in an illegal vertex was accepted
	status, err = virtuousXChainAPI.GetTxStatus(nonConflictId)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get transaction status for non-conflicting transaction.")
	}
	logrus.Infof("Status of CreateAssetTx: %s is %s", nonConflictId, status)
	// If the transaction was Accepted, the test should fail because virtuous nodes should not issue the vertex and
	// the underlying transactions into consensus
	if status == choices.Accepted {
		return stacktrace.Propagate(err, fmt.Sprintf("Expected status of non-conflicting transaction issued in bad vertex to be Processing, but found %s", status))
	}
	return nil
}
