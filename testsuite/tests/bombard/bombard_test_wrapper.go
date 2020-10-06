package bombard

import (
	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"
	"time"

	avalancheNetwork "github.com/ava-labs/avalanche-testing/avalanche/networks"
	avalancheService "github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigID       networks.ConfigurationID = "normal-config"
	additionalNode1ServiceID                          = "additional-node-1"
	additionalNode2ServiceID                          = "additional-node-2"
	seedAmount                                        = int64(50000000000000)
	stakeAmount                                       = int64(30000000000000)
)

// StakingNetworkBombardTest funds individual clients with a starting UTXO for each
// and then creates a string of transactions to send to each one based off of the original UTXO.
// Then it adds two nodes to ensure that they can bootstrap the new data on the X chain.
type StakingNetworkBombardTest struct {
	ImageName         string
	NumTxs            uint64
	TxFee             uint64
	AcceptanceTimeout time.Duration
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkBombardTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(avalancheNetwork.TestAvalancheNetwork)
	bootServiceIDs := castedNetwork.GetAllBootServiceIDs()
	clients := make([]*apis.Client, 0, len(bootServiceIDs))
	for serviceID := range bootServiceIDs {
		avalancheClient, err := castedNetwork.GetAvalancheClient(serviceID)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get Avalanche Client for boot node with serviceID: %s.", serviceID))
		}
		clients = append(clients, avalancheClient)
	}

	// Execute the bombard test to issue [NumTxs] to each node
	executor := NewBombardExecutor(clients, test.NumTxs, test.TxFee, test.AcceptanceTimeout, 1)
	logrus.Infof("Executing bombard test...")
	if err := executor.ExecuteTest(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Bombard Test Failed."))
	}

	logrus.Infof("Bombard test completed successfully.")
	logrus.Infof("Adding two additional nodes and waiting for them to bootstrap...")
	// Add two additional nodes to ensure that they can successfully bootstrap the additional data
	availabilityChecker1, err := castedNetwork.AddService(normalNodeConfigID, additionalNode1ServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add %s to the network.", additionalNode1ServiceID))
	}
	availabilityChecker2, err := castedNetwork.AddService(normalNodeConfigID, additionalNode2ServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add %s to the network.", additionalNode2ServiceID))
	}

	// Wait for the nodes to finish bootstrapping
	if err = availabilityChecker1.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to wait for startup of %s.", additionalNode1ServiceID))
	}
	logrus.Infof("Node1 finished bootstrapping.")
	if err = availabilityChecker2.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to wait for startup of %s.", additionalNode2ServiceID))
	}
	logrus.Infof("Node2 finished bootstrapping.")
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkBombardTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Add config for a normal node, to add an additional node during the test
	desiredServices := make(map[networks.ServiceID]networks.ConfigurationID)
	serviceConfigs := make(map[networks.ConfigurationID]avalancheNetwork.TestAvalancheNetworkServiceConfig)
	serviceConfigs[normalNodeConfigID] = *avalancheNetwork.NewTestAvalancheNetworkServiceConfig(
		true,
		avalancheService.DEBUG,
		test.ImageName,
		2,
		2,
		2*time.Second,
		make(map[string]string),
	)

	return avalancheNetwork.NewTestAvalancheNetworkLoader(
		true,
		test.ImageName,
		avalancheService.DEBUG,
		2,
		2,
		test.TxFee,
		2*time.Second,
		serviceConfigs,
		desiredServices,
	)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test StakingNetworkBombardTest) GetExecutionTimeout() time.Duration {
	return 10 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test StakingNetworkBombardTest) GetSetupBuffer() time.Duration {
	return 2 * time.Minute
}
