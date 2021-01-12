package managedasset

import (
	"strconv"
	"time"

	avalancheNetwork "github.com/ava-labs/avalanche-testing/avalanche/networks"
	avalancheService "github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigID       networks.ConfigurationID = "normal-config"
	additionalNode1ServiceID networks.ServiceID       = "additional-node-1"
	additionalNode2ServiceID networks.ServiceID       = "additional-node-2"
)

// ManagedAssetTest implements the testsuite.Test interface and tests the
// functionality of managed assets
type ManagedAssetTest struct {
	ImageName string
}

// NewManagedAssetTest returns a new Kurtosis Test
func NewManagedAssetTest(imageName string) testsuite.Test {
	return &ManagedAssetTest{
		ImageName: imageName,
	}
}

// Run implements the Kurtosis Test interface
func (test ManagedAssetTest) Run(network networks.Network, context testsuite.TestContext) {
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

	// Execute the test
	executor := Executor{
		clients:           clients,
		acceptanceTimeout: 3 * time.Second,
		epochDuration:     5 * time.Second, // this should match the CLI arg passed in GetNetworkLoader
	}
	logrus.Infof("Executing managed asset test...")
	if err := executor.Execute(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Managed asset test failed."))
	}

	logrus.Infof("Managed asset test completed successfully.")
	logrus.Infof("Adding two additional nodes and waiting for them to bootstrap...")
	// // Add two additional nodes to ensure that they can successfully bootstrap the additional data
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
func (test ManagedAssetTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Add config for a normal node, to add an additional node during the test
	desiredServices := make(map[networks.ServiceID]networks.ConfigurationID)
	serviceConfigs := make(map[networks.ConfigurationID]avalancheNetwork.TestAvalancheNetworkServiceConfig)
	now := time.Now().Unix() - 1
	nowStr := strconv.Itoa(int(now))

	serviceConfig := *avalancheNetwork.NewTestAvalancheNetworkServiceConfig(
		true,                   // is staking
		avalancheService.DEBUG, // log level
		test.ImageName,         // image name
		2,                      // snow quorum size
		3,                      // snow sample size
		2*time.Second,          // network initial timeout
		map[string]string{
			"snow-epoch-first-transition": nowStr,
			"snow-epoch-duration":         "5s",
		}, // additional CLI args
	)

	serviceConfigs[normalNodeConfigID] = serviceConfig

	return avalancheNetwork.NewTestAvalancheNetworkLoader(
		true,
		1, // tx fee
		serviceConfig,
		serviceConfigs,
		desiredServices,
	)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test ManagedAssetTest) GetExecutionTimeout() time.Duration {
	return 10 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test ManagedAssetTest) GetSetupBuffer() time.Duration {
	return 2 * time.Minute
}
