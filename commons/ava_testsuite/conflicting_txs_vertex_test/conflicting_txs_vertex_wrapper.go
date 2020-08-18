package conflicting_txs_vertex_test

import (
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_networks"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_services"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigID          networks.ConfigurationID = "normal-config"
	byzantineConfigID           networks.ConfigurationID = "byzantine-config"
	byzantineUsername                                    = "byzantine_gecko"
	byzantinePassword                                    = "byzant1n3!"
	byzantineBehavior                                    = "byzantine-behavior"
	conflictingTxVertexBehavior                          = "conflicting-txs-vertex"
	stakerUsername                                       = "staker_gecko"
	stakerPassword                                       = "test34test!23"
	byzantineNodeServiceID                               = "byzantine-node"
	normalNodeServiceID                                  = "virtuous-node"
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

	byzantineClient, err := castedNetwork.GetGeckoClient(byzantineNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get byzantine client."))
	}
	virtuousClient, err := castedNetwork.GetGeckoClient(normalNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get virtuous client."))
	}
	executor := NewConflictingTxsVertexExecutor(virtuousClient, byzantineClient)
	if err := executor.ExecuteTest(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Conflicting Transactions Vertex Test failed."))
	}
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkConflictingTxsVertexTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Provision a byzantine and normal node
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{}
	desiredServices[byzantineNodeServiceID] = byzantineConfigID
	desiredServices[normalNodeServiceID] = normalNodeConfigID

	return getByzantineNetworkLoader(desiredServices, test.ByzantineImageName, test.NormalImageName)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test StakingNetworkConflictingTxsVertexTest) GetExecutionTimeout() time.Duration {
	return 2 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
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
		normalNodeConfigID: *ava_networks.NewTestGeckoNetworkServiceConfig(
			true,
			ava_services.LOG_LEVEL_DEBUG,
			normalImageName,
			2,
			2,
			make(map[string]string),
		),
		byzantineConfigID: *ava_networks.NewTestGeckoNetworkServiceConfig(
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
