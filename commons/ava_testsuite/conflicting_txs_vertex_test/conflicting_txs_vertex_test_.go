package conflicting_txs_vertex_test

import (
	"fmt"
	"time"

	"github.com/ava-labs/gecko/snow/choices"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_networks"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigId          networks.ConfigurationID = "normal-config"
	byzantineConfigId           networks.ConfigurationID = "byzantine-config"
	byzantineUsername                                    = "byzantine_gecko"
	byzantinePassword                                    = "byzant1n3!"
	byzantineBehavior                                    = "byzantine-behavior"
	conflictingTxVertexBehavior                          = "conflicting-txs-vertex"
	stakerUsername                                       = "staker_gecko"
	stakerPassword                                       = "test34test!23"
	byzantineNodeServiceId                               = "byzantine-node"
	normalNodeServiceId                                  = "virtuous-node"
	seedAmount                                           = int64(50000000000000)
	stakeAmount                                          = int64(30000000000000)
)

// ================ Byzantine Test - Conflicting Transactions in a Vertex Test ===================================
// StakingNetworkConflictingTxsVertexTest implements the Test interface
type StakingNetworkConflictingTxsVertexTest struct {
	ByzantineImageName string
	NormalImageName    string
}

// Issue conflicting transactions to the byzantine node to be issued into a vertex
// The byzantine node should mark them as accepted when it issues them into a vertex.
// Once the transactions are issued, verify the byzantine node has marked them as accepted
// Virtuous nodes should drop the vertex without issuing it the vertex or its transactions
// into consensus.
// As a result both the virtuous and rogue transactions within the vertex should stay stuck
// in processing.
func (test StakingNetworkConflictingTxsVertexTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.TestGeckoNetwork)

	byzantineClient, err := castedNetwork.GetGeckoClient(byzantineNodeServiceId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get byzantine client."))
	}

	byzantineXChainAPI := byzantineClient.XChainAPI()
	// TODO switch to test vectors or come up with method to reliably generate conflicting transactions
	// how to create a conflicting transaction???
	// test vector create asset tx and conflicting transactions
	createAssetTx := []byte{1}
	conflictingTx1 := []byte{2}
	conflictingTx2 := []byte{3}

	nonConflictId, err := byzantineXChainAPI.IssueTx(createAssetTx)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to issue first transaction to byzantine node."))
	}
	conflictId1, err := byzantineXChainAPI.IssueTx(conflictingTx1)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to issue second transaction to byzantine node."))
	}
	conflictId2, err := byzantineXChainAPI.IssueTx(conflictingTx2)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to issue third transaction to byzantine node."))
	}

	logrus.Infof("Issued three transactions to Byzantine Node with IDs: %s, %s, %s", nonConflictId, conflictId1, conflictId2)

	// Confirm the byzantine node Accepted the transactions
	// Note: The byzantine behavior is to batch the pending transactions into a vertex as soon as it detects a conflict.
	// It should try to accept each transaction before PushQuery-ing the vertex to other nodes to signal to this test
	// controller that the vertex was successfully issued
	status, err := byzantineXChainAPI.GetTxStatus(nonConflictId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", nonConflictId)))
	}
	if status != choices.Accepted {
		context.Fatal(stacktrace.Propagate(err, fmt.Sprintf("Transaction: %s was not accepted, status: %s", nonConflictId, status)))
	}

	logrus.Infof("Status of non-conflict transactions on byzantine node is: %s", status)

	conflictStatus1, err := byzantineXChainAPI.GetTxStatus(conflictId1)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", conflictId1)))
	}

	logrus.Infof("Status of conflict tx1: %s on byzantine node is: %s", conflictId1, conflictStatus1)

	conflictStatus2, err := byzantineXChainAPI.GetTxStatus(conflictId2)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", conflictId2)))
	}

	logrus.Infof("Status of conflict tx2: %s on byzantine node is: %s", conflictId2, conflictStatus2)

	// Byzantine node should try to accept both conflicting transactions, but will fail to accept one due to the missing UTXO
	// after the other consumes it.
	if conflictStatus1 != choices.Accepted && conflictStatus2 != choices.Accepted {
		context.Fatal(fmt.Errorf("Byzantine node did not accept either of the conflicting transactions, status1: %s. status2: %s", conflictStatus1, conflictStatus2))
	}

	// The issued vertex should be dropped completely, so the virtuous nodes should drop the vertex
	// and never issue the transactions into consensus.
	// Note: since the transactions will be parsed in the process, we expect the status to be "Processing" not "Rejected" or "Unknown"
	virtuousClient, err := castedNetwork.GetGeckoClient(normalNodeServiceId)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get virtuous client."))
	}

	// We issue a vertex from the virtuous node to see if it builds on invalid vertex
	// This is meant to remove the need to wait an arbitrary amount of time to see if the vertex gets accepted
	// and instead confirm the valid transaction as a measure of the time to finality before checking if
	// the transactions that should have been dropped were in fact dropped successfully.
	// TODO move to test vector
	virtuousXChainAPI := virtuousClient.XChainAPI()
	virtuousCreateAssetTx := []byte{4}
	virtuousSpendTx := []byte{5}

	// Ignore the TxID of this because it should be accepted immediately after entering consensus
	_, err = virtuousXChainAPI.IssueTx(virtuousCreateAssetTx)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to issue virtuous create asset transaction after issuing illegal vertex from byzantine node."))
	}
	virtuousSpendTxId, err := virtuousXChainAPI.IssueTx(virtuousSpendTx)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to issue virtuous transaction spending created asset after issuing byzantine vertex"))
	}

	for {
		status, err := virtuousXChainAPI.GetTxStatus(virtuousSpendTxId)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get virtuous transactions status from virtuous node"))
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
		context.Fatal(stacktrace.Propagate(err, "Failed to get transaction status for non-conflicting transaction."))
	}
	logrus.Infof("Status of CreateAssetTx: %s is %s", nonConflictId, status)
	// If the transaction was Accepted, the test should fail because virtuous nodes should not issue the vertex and
	// the underlying transactions into consensus
	if status == choices.Accepted {
		context.Fatal(stacktrace.Propagate(err, fmt.Sprintf("Expected status of non-conflicting transaction issued in bad vertex to be Processing, but found %s", status)))
	}
}

func (test StakingNetworkConflictingTxsVertexTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Provision a byzantine and normal node
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{}
	desiredServices[byzantineNodeServiceId] = byzantineConfigId
	desiredServices[normalNodeServiceId] = normalNodeConfigId

	return getByzantineNetworkLoader(desiredServices, test.ByzantineImageName, test.NormalImageName)
}

func (test StakingNetworkConflictingTxsVertexTest) GetExecutionTimeout() time.Duration {
	return 2 * time.Minute
}

func (test StakingNetworkConflictingTxsVertexTest) GetSetupBuffer() time.Duration {
	return 2 * time.Minute
}

// =============== Helper functions =============================

/*
Args:
	desiredServices: Mapping of service_id -> configuration_id for all services *in addition to the boot nodes* that the user wants
*/
func getByzantineNetworkLoader(desiredServices map[networks.ServiceID]networks.ConfigurationID, byzantineImageName string, normalImageName string) (networks.NetworkLoader, error) {
	serviceConfigs := map[networks.ConfigurationID]ava_networks.TestGeckoNetworkServiceConfig{
		normalNodeConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(
			true,
			ava_services.LOG_LEVEL_DEBUG,
			normalImageName,
			2,
			2,
			make(map[string]string),
		),
		byzantineConfigId: *ava_networks.NewTestGeckoNetworkServiceConfig(
			true,
			ava_services.LOG_LEVEL_DEBUG,
			byzantineImageName,
			2,
			2,
			map[string]string{byzantineBehavior: conflictingTxVertexBehavior},
		),
	}
	logrus.Debugf("Byzantine Image Name: %s", byzantineImageName)
	logrus.Debugf("Normal Image Name: %s", normalImageName)

	return ava_networks.NewTestGeckoNetworkLoader(
		true,
		normalImageName,
		ava_services.LOG_LEVEL_DEBUG,
		2,
		2,
		serviceConfigs,
		desiredServices,
	)
}
