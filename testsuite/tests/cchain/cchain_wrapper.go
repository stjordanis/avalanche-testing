package cchain

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
	normalNodeConfigID       networks.ConfigurationID = "normal-config"
	additionalNode1ServiceID networks.ServiceID       = "additional-node-1"
	additionalNode2ServiceID networks.ServiceID       = "additional-node-2"
)

// Test runs a series of basic C-Chain tests on a network of
// virtuous nodes
type Test struct {
	ImageName      string
	NumTxs         int
	NumTxLists     int
	TxFee          uint64
	RequestTimeout time.Duration
}

// NewVirtuousCChainTest ...
func NewVirtuousCChainTest(imageName string, numTxs int, numTxLists int, txFee uint64, requestTimeout time.Duration) testsuite.Test {
	return &Test{
		ImageName:      imageName,
		NumTxs:         numTxs,
		NumTxLists:     numTxLists,
		TxFee:          txFee,
		RequestTimeout: requestTimeout,
	}
}

// Run implements the Kurtosis Test interface
func (test Test) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(avalancheNetwork.TestAvalancheNetwork)
	bootServiceIDs := castedNetwork.GetAllBootServiceIDs()
	clients := make([]*avalancheService.Client, 0, len(bootServiceIDs))
	for serviceID := range bootServiceIDs {
		avalancheClient, err := castedNetwork.GetAvalancheClient(serviceID)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get Avalanche Client for boot node with serviceID: %s.", serviceID))
		}
		clients = append(clients, avalancheClient)
	}

	logrus.Infof("Executing basic consecutive transactions test.")
	if err := NewBasicTransactionThroughputTest(clients, 5, 1000).ExecuteTest(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Basic transaction test failed"))
	}
	logrus.Infof("Basic transaction test completed successfully.")

	logrus.Infof("Executing basic ETH API test.")
	if err := NewEthAPIExecutor(clients[0].CChainEthAPI()).ExecuteTest(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Basic API test failed"))
	}
	logrus.Infof("Basic ETH API test completed successfully.")

	logrus.Infof("Executing C-Chain Atomic Workflow Test.")
	if err := CreateAtomicWorkflowTest(clients[0], test.TxFee).ExecuteTest(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "C-Chain Atomic Workflow Test Failed."))
	}
	logrus.Infof("C-Chain Atomic Workflow Test completed successfully.")

	logrus.Infof("C-Chain Tests completed successfully.")
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
func (test Test) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Add config for a normal node, to add an additional node during the test
	desiredServices := make(map[networks.ServiceID]networks.ConfigurationID)
	serviceConfigs := make(map[networks.ConfigurationID]avalancheNetwork.TestAvalancheNetworkServiceConfig)
	serviceConfig := *avalancheNetwork.NewDefaultAvalancheNetworkServiceConfig(test.ImageName)
	serviceConfigs[normalNodeConfigID] = serviceConfig

	return avalancheNetwork.NewDefaultAvalancheNetworkLoader(
		true,            // Staking network
		test.TxFee,      // Network wide transaction fee
		serviceConfig,   // Config for the bootstrap nodes
		serviceConfigs,  // Service Configurations
		desiredServices, // Services to start on network configuration
	)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test Test) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test Test) GetSetupBuffer() time.Duration {
	return 2 * time.Minute
}
