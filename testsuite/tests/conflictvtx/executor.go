package conflictvtx

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche_client/apis"
	"github.com/ava-labs/avalanche-testing/testsuite/tester"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/formatting"
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

	if err := cb58.FromString("111118ErNmgrybbV39jBv1GAdzTWsC1Smd3WxkcTSe6uPvaxAtJmcCWGgRT4mj6V6qdoK1iZzGeUrqujSedUE4rF5sjryxy5uVkQo6CMit9GtzyyQmhtRMjUcNMhT9qgJ92EQYzP2PWxjHQkc3SbYxtBttB8ccMBFSVHh8LipqWCk9FkK5vjoTmvC6YtRNAdg81Ye4qK8MPEY38U9N2ckgCRxjkZg7ZyHUnh1vACsdGcyQgDhWmwhUgNznAEXWVvpnbBznyhicbXR3mwsrsck16xrzHwr96zVugZzEeikChAndTjaJoa2GCF4jqv5zsHd5PQcoQQk4QdgrtjxFbzXBaJW54MrELgenAHFSzq4nrToyxsdm4SA6pUNa2bGDWnouskrVHHdvXcVueE7dp5PBcuix1nD2VR8VZzTi9ZKDJ79vQv8pcCTE1ZUEPxkAsQFTdv9XJVCikThtKLXxrDdvBybuk9iUbqjrzRHAuY93VoZ4evjFD2z6Wtrhd9DRoLcyAX8Ft"); err != nil {
		return stacktrace.Propagate(err, "Problem parsing create asset tx")
	}
	createAssetTx := cb58.Bytes
	if err := cb58.FromString("11111111Wc1cPy8nDRbzTRYwGxcqaSC5eCqSmpK855uz6HP8FQZMXSja7nmUmRSqu54Y6xjnPr23fbqqQDmf4GU3pXfmUHXfYSiCRqBG1TRxnmEf58Va8nYijMcSMjdWGgaoXyMvuh6Rhhz4YXzFbTMuAMrwsFuXGKZ5wD1YKk4q5JXpppeJWd3uxrN4iQqcHeVkAWyQY7h6suuqet6gtKPYoRDRj9w6aZmerHhw7BCepbtSdatFoaVj7aAdV6YMBnrF7B4rRnWUH7DHLnorAqWZCVgVKShZdMNYaX5HUGJnqAFCMVt6ZhgvyDJT3JBWLRKEVXVGe34ziE9XHBMRoUjHforJeLbMMuQVUBvgbKVJC7eWnjCNh5xTGJoaPyGWUht8L6Ws2T6aFEx97sysJmxk6V318wDsWFmvhhgLtr4eYQmQt6xwfVA1DVFjm5tHMfFYo5RWD2XDrEqsteRBVi6RMnfDcE7h95dKDo2Pd4ZDpeEEbsZRtC3oJoHA4Cq9HMxdmbYH83PRvvez4wopvHujmSPmqkiHhaDaCdPx7XTdh95gn4u2nPQCqUjHhbad5RYNPxP1CcsSrJkZ6PSgBkkoNXJazeUVmXoniBVpXKqZtsgnsUfVcRGsuoZ7g5qaRSUMqTLR7GUY7yg1DXg8VSwFeTEkvVsvb2AM4E2BWEhqZwdj4qgAg8kAfuDDz8XMzpT36ie8UirEASyjbLdgrB5JBsj3mF1e3iX42DNAFU6awrtjroRaJ3gM9r75ALDuwRKjrCPYp7aJ1ioCRzgfSUPaP33wJZVh98ZJp8Tv3Eq4k6VMcWMVor41JZLskivuWN9gRBqhQRim3Fvhm"); err != nil {
		return stacktrace.Propagate(err, "Problem parsing first send tx")
	}
	conflictingTx1 := cb58.Bytes

	if err := cb58.FromString("11111111Wc1cPy8nDRbzTRYwGxcqaSC5eCqSmpK855uz6HP8FQZMXSja7nmUmRSqu54Y6xjnPr23fbqqQDmf4GU3pXfmUHXfYSiCRqBG1TRxnmEf58Va8nyeWyVPwGH2wV6EECrLQ5UHHdVCn7z5SmpfVMp7R6GFiK8sq7evSEi1T2xPzBMwATe4RtJeaHSKzZzzK2aMDLR9ojbnddBMCxwczH596n8PN19b9Woc2uyTLZzcjWZCfagtHoNyZVX9Xd4ZkNiAMTH3v9FmNZaMWYQraxjaSozvKvCzyd1ESrKMKdEHVf1jK7Ucfs7RxEx4ApRF2deC5P3thaNLaescZWeu3vnrnYvHr5Faz58HRSHfyxhcRJtukUrSFA2rfkba4wq8wHuUeVURMGftgmm3SKg7FWJrdoouyAfWkCZa1w6cUAVYxbp319v2d2xfibJiiTptVzF4m1zjVkwsU82Md128xNmPJ7uXKJcdh9GcS3HRho17DvzjEXECwPh9qN6oynXs5LZEjfNCU5DHVACNutq2hyubgLd7oozMUGKfqCNmn6aBTxNcTMVt6nyANkckaYGEJGs84gbLtfxFG13Uwqgck6ciMPbgdepwWU3eNZzLXGMw8mnBjEQx6yFy18HMENXSCjmjCchhSNJ2i1c9MJU5TjDYCtTWegVcGvEJrgFj36rjHRA6ACNtYsiW1Fychn2fnuqMuvBcyHo5wrNo7UDq54vGbfLpTxkxtQfHeBwN5bufFUGurhU2BxwsLpXsqBknXow7ScFP9HqYM3HNTFUuSwoyTa4e1Hf99KunMuXXT97eS3FTxzoNY9rRmSrGBjiBKFeyQviSqWyay"); err != nil {
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
	if err := cb58.FromString("11111NYyFi7CerEDNe3SKAsqAtuKsw6cyu4ysavdJjsAZs1EkYFJsYZ8AhV47ayaxzq4MUfYQhp3vx9dsD1QyYATTiiR1hZDfUcp3bAJ4vCnR5YrtYLH6quiKVLDbnXpVW7YJyPacmUM8V3EcVEUTNWNxt25urNLQ39fdrVbtnWvEkcY2hSSNH2y1ayYwUgkvtxVnmwGMQBXgY1gYW9JZsVTaBsgZ6W5xEdPZAiuqrqdRxNzhEEDuPE1kh3ncex9buu5b6ben5PvsjjBZaChr7k2h7FLQgi5JJBamkskZaVEdckNDXHnHoA5L9ApLyWsxxut8i6BhjYrZBdKYFZHEcm2U21ciSoxgv9j6j1AaDDMjAbpGJ8DxBQYBaNyxWNZB2zNcP4nVsduoQUbmqAsZHGtCohMf1HFpHmbNWMs7ekzd9nxP13cSAqcmibJyZC9gurEMc68vijm17MTHHM8tVBPUFjj3Uy6rtfXZ3AxAs1ELPjMztfYb6fN4BYAKkAodaAP"); err != nil {
		return stacktrace.Propagate(err, "Failed to parse virtuous create asset tx")
	}
	virtuousCreateAssetTx := cb58.Bytes

	// Ignore the TxID of this because it should be accepted immediately after entering consensus
	virtuousTxID, err := virtuousXChainAPI.IssueTx(virtuousCreateAssetTx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue virtuous create asset transaction after issuing illegal vertex from byzantine node.")
	}

	for {
		status, err := virtuousXChainAPI.GetTxStatus(virtuousTxID)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get virtuous transactions status from virtuous node")
		}
		if status == choices.Accepted {
			logrus.Infof("Accepted virtuous transaction with ID: %s", virtuousTxID)
			break
		} else {
			logrus.Infof("Waiting for transaction with ID: %s to be accepted", virtuousTxID)
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
