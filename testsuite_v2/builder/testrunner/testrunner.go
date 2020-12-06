package testrunner

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"
)

// TestRunner is the implementation of the interface used to run tests on
type TestRunner struct {
	name             string
	network          func() (networks.NetworkLoader, error)
	test             func(network networks.Network, context testsuite.TestContext)
	executionTimeout time.Duration
	bufferTimeout    time.Duration
}

// NewTestRunner creates a new instance
func NewTestRunner(
	name string,
	network func() (networks.NetworkLoader, error),
	test func(network networks.Network, context testsuite.TestContext),
	executionTimeout time.Duration,
	bufferTimeout time.Duration,
) *TestRunner {
	return &TestRunner{
		name:             name,
		network:          network,
		test:             test,
		executionTimeout: executionTimeout,
		bufferTimeout:    bufferTimeout,
	}
}

// NOTE: if Go had generics, 'network' would be a parameterized type representing the network that this test consumes
// as produced by the NetworkLoader
/*
	Runs test logic against the given network, with failures reported using the given context.

	Args:
		network: A user-defined representation of the network. NOTE: Because Go doesn't have generics, this will need to
			be casted to the appropriate type.
		context: The test context, which is the user's tool for making test assertions.
*/
func (tr *TestRunner) Run(network networks.Network, context testsuite.TestContext) {
	logrus.Infof("%s is started...\n", tr.name)
	tr.test(network, context)
	logrus.Infof("%s is finished...\n", tr.name)
}

// Gets the network loader that will be used to spin up the test network that the test will run against
func (tr *TestRunner) GetNetworkLoader() (networks.NetworkLoader, error) {
	return tr.network()
}

/*
	The amount of time the test's `Run` method will be allowed to execute for before it's killed and the test
		is marked as failed. This does NOT include the time needed to do pre-test setup or post-test teardown,
		which is handled by `GetSetupBuffer`. The total amount of time a test (with setup & teardown) is allowed
		to run for = GetExecutionTimeout + GetSetupBuffer.
*/
func (tr *TestRunner) GetExecutionTimeout() time.Duration {
	return tr.executionTimeout
}

/*
	How long the test will be given to do the pre-execution setup and post-setup teardown before the test will be
		hard-killed. The total amount of time a test (with setup & teardown) is allowed to run
		for = GetExecutionTimeout + GetSetupBuffer.
*/
func (tr *TestRunner) GetSetupBuffer() time.Duration {
	return tr.bufferTimeout
}
