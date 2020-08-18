package conflicting_txs_vertex_test

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/ava-labs/gecko/utils/formatting"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type executor struct {
	virtuousClient  *apis.Client
	byzantineClient *apis.Client
}

// NewConflictingTxsVertexExecutor ...
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

	cb58 := formatting.CB58{}
	if err := cb58.FromString("111115P99NeFpAqLusu4fw9f8yqU89bnDn4yKjvvBKZPdfzTUHkjdh56Bh5SJhM83LWcPp6nzuDtttQVG6qaStJQs5uG5tgt6WUWAopGjZ149Vgtz9KmuQHXxQvF7Gb5d9kJ3ebRJSU9yNeygeWTndyDbqoyEFkCzxthdZuLs1HpQZZfpDztDdhLDkop2F"); err != nil {
		return stacktrace.Propagate(err, "Problem parsing create asset tx")
	}
	createAssetTx := cb58.Bytes
	if err := cb58.FromString("111111117RZoLdg3oME39MoTLGqVQHyHKz7MdYuVgZC1Luf3YSBEhhDSi5RfdXEDuQvd5YwujF2wCi1YRA8vSMsf1LHwvuWEZS4ct1uQ7KHQKYGR6c99PM8tyRCNaAaGLLtDLF6N4Yjrqfpqji7JoRafz9ytHGN7AjWHcDZUcQ5dRjpwgbCSu3A1My355B5niXyBMv2Eg4A3z7uLfrCoc8QvXv8DThhpht6FV43cvLcM6xyL6eALsBekp946n3d5evMsdp7n4xi1Rc6rfRctnm5Nqx3peFvZKxjDaJLXMBF35xZjnnjDtiHTPX6VFMiiwBvkdmgo1aTkEMJMYm249XvqC5R7bjFU8bYQ9orsFQGV2QvVHazCxyHJGJ1kqiMhdUV5HKGGZfHMJS4bcRqsiF2CDr1"); err != nil {
		return stacktrace.Propagate(err, "Problem parsing first send tx")
	}
	conflictingTx1 := cb58.Bytes

	if err := cb58.FromString("111111117RZoLdg3oME39MoTLGqVQHyHKz7MdYuVgZC1Luf3YSBEhhDSi5RfdXEDuQvd5YwujF2wCi1YRA8vSMsf1LHwvuWEZS4ct1uQ7KHQKYGR6c99PM8tyRCNaAaGLLtDLF6N4YjrqfH5EDCRFhJFfkywrVMyMTMKRc6RC3nwEzjDcKWhHyS9w4cc9i2bT4gTK7hj3q4bHcYuwEnY2tBPtsgXaSVWb4XZCFLaJgkNxpwP3nxiJr87syt9TWMRgzo5sNq5vjiez8bDqhsFy9nyRrPZ3UBBQpkj9YG1g5vuNE2oX2umji7A89NxX6yV4djvQ8CK3HbJpLhnazMxYdvhzVVFjFKrkTVgr8xkadi2s6e9AMdhxxVv4ovpG2pFxFjnF8aZasdzqXgyZDuoEhQ1Zn1"); err != nil {
		return stacktrace.Propagate(err, "Problem parsing second send tx")
	}
	conflictingTx2 := cb58.Bytes

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
	if err := cb58.FromString("111115P99NeFpAqLusu4fw9f8yqU89bnDn4yKjvvBKZPdfzTUHkjdh56Bh5SJhM83LWcPp6nzuDtttQVG6qaStJQs5uG5tgt6WUWAopGjZ149Vgtz9KmuQHXxQvF7Gb3VSJTiEmsmZ5Eb3q7ergPFFxZbqVmGrLCJmnu7gYDjfisxR8WHuwCdjyGF2E1Y9"); err != nil {
		return stacktrace.Propagate(err, "Failed to parse virtuous create asset tx")
	}
	virtuousCreateAssetTx := cb58.Bytes
	if err := cb58.FromString("111111117RZoLdg3oME39MoTLGqVQHyHKz7MdYuVgZC1Luf3YSBEhhDSi5RfBUbmsdPHhbY9CkpPY6JRtqWprtt8upHXBHhN4hPx3ko3vdsn7wxhi1HthuLQTixEbsV2fCcs6Po4rMYmcMM9uvMztwf1mUQREgjk674e27VKsQDh6BcDng9wiXjt1AbgBt3rbaLgwLh1fRQas81VH6hRQEqgCkNqwYt7nVWGLN8YEbXbYbkX4KGiYrc5n66A6mdJBzqh111t1KSRRZd3hbtzPunwx2QXa5RjogQYeWUsiAoscvVknYzVYuMcX9mqZnr163JrFVxp3Wwv5HK41uu8kFQUzofoHkmLL1PMS5ASDLEzeYgyX2oqjU6rxhs6y5ayiHnfnpehKMLYHWfFj2fkgmeEgV3"); err != nil {
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
