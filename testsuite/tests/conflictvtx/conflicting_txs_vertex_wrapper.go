package conflictvtx

import (
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	avalancheNetwork "github.com/ava-labs/avalanche-testing/avalanche/networks"
	avalancheService "github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigID          networks.ConfigurationID = "normal-config"
	byzantineConfigID           networks.ConfigurationID = "byzantine-config"
	byzantineBehavior                                    = "byzantine-behavior"
	conflictingTxVertexBehavior                          = "conflicting-txs-vertex"
	byzantineNodeServiceID                               = "byzantine-node"
	normalNodeServiceID                                  = "virtuous-node"
)

// StakingNetworkConflictingTxsVertexTest creates a byzantine node to issue conflicting transactions into a single
// vertex. It then checks to ensure that the byzantine node has accepted these transactions, while the virtuous nodes
// drop the vertex.
type StakingNetworkConflictingTxsVertexTest struct {
	ByzantineImageName string
	NormalImageName    string
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkConflictingTxsVertexTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(avalancheNetwork.TestAvalancheNetwork)

	byzantineClient, err := castedNetwork.GetAvalancheClient(byzantineNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get byzantine client."))
	}
	virtuousClient, err := castedNetwork.GetAvalancheClient(normalNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get virtuous client."))
	}
	executor := NewConflictingTxsVertexExecutor(virtuousClient, byzantineClient)
	logrus.Infof("Executing conflicting transaction vertex test...")
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
	serviceConfigs := map[networks.ConfigurationID]avalancheNetwork.TestAvalancheNetworkServiceConfig{
		normalNodeConfigID: *avalancheNetwork.NewTestAvalancheNetworkServiceConfig(
			true,
			avalancheService.DEBUG,
			normalImageName,
			2,
			2,
			2*time.Second,
			make(map[string]string),
		),
		byzantineConfigID: *avalancheNetwork.NewTestAvalancheNetworkServiceConfig(
			true,
			avalancheService.DEBUG,
			byzantineImageName,
			2,
			2,
			2*time.Second,
			map[string]string{byzantineBehavior: conflictingTxVertexBehavior},
		),
	}
	logrus.Debugf("Byzantine Image Name: %s", byzantineImageName)
	logrus.Debugf("Normal Image Name: %s", normalImageName)

	return avalancheNetwork.NewTestAvalancheNetworkLoader(
		true,
		normalImageName,
		avalancheService.DEBUG,
		2,
		2,
		nil,
		1000000,
		2*time.Second,
		serviceConfigs,
		desiredServices,
	)
}
