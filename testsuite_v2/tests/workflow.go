package tests

import (
	"time"

	top "github.com/ava-labs/avalanche-testing/testsuite_v2/builder/topology"

	"github.com/ava-labs/avalanche-testing/testsuite_v2/builder/chainhelper"

	"github.com/sirupsen/logrus"

	"github.com/palantir/stacktrace"

	"github.com/ava-labs/avalanche-testing/testsuite_v2/builder/network"
	"github.com/ava-labs/avalanche-testing/utils/constants"

	"github.com/ava-labs/avalanchego/utils/units"

	"github.com/ava-labs/avalanche-testing/testsuite_v2/builder/testrunner"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"
)

const (
	genesisUsername   = "genesis"
	genesisPassword   = "MyNameIs!Jeff"
	stakerUsername    = "staker"
	stakerPassword    = "test34test!23"
	delegatorUsername = "delegator"
	delegatorPassword = "test34test!23"
	totalAmount       = 10 * units.KiloAvax
	seedAmount        = 5 * units.KiloAvax
	stakeAmount       = 3 * units.KiloAvax
	delegatorAmount   = 3 * units.KiloAvax
)

const (
	regularNodeServiceID   networks.ServiceID = "validator-node"
	delegatorNodeServiceID networks.ServiceID = "delegator-node"
)

// Workflow (it's disabled) demos the test of the workflow - copy of the ../testsuite/tests/workflow test
// mostly useful for showing how to manually build the network and the topology
func Workflow(avalancheImage string) *testrunner.TestRunner {

	// create the nodes
	stakerNode := network.NewNode("validator-node").
		Image(avalancheImage).
		SnowConf(2, 2)

	delegatorNode := network.NewNode("delegator-node").
		Image(avalancheImage).
		SnowConf(2, 2)

	// creates the network + adds the Nodes
	testNetwork := network.New().
		IsStaking(true).
		Image(avalancheImage).
		SnowSize(2, 2).
		AddNode(stakerNode).
		AddNode(delegatorNode)

	// timeout
	timeout := 3 * time.Minute

	// TODO drop this down when the availability checker doesn't have a sleep (because we spin up a bunch of nodes before the test starts executing)
	timeoutBuffer := 3 * time.Minute

	// the actual test
	test := func(network networks.Network, context testsuite.TestContext) {

		// builds the topology of the test
		topology := top.New(network, &context)
		topology.
			AddNode("validator-node", stakerUsername, stakerPassword).
			AddNode("delegator-node", delegatorUsername, delegatorPassword).
			AddGenesis("validator-node", genesisUsername, genesisPassword)

		// creates a genesis and funds the X addresses of the nodes
		topology.Genesis().
			FundXChainAddresses([]string{
				topology.Node("validator-node").XAddress,
				topology.Node("delegator-node").XAddress,
			},
				totalAmount,
			)

		// sets the nodes to validators and delegators
		topology.Node("validator-node").BecomeValidator(totalAmount, seedAmount, stakeAmount)
		topology.Node("delegator-node").BecomeDelegator(seedAmount, stakeAmount, topology.Node("validator-node").NodeID)

		// after setup we want to test moving amounts from P to X Chain and back
		stakerNode := topology.Node("validator-node")

		exportTxID, err := stakerNode.GetClient().PChainAPI().ExportAVAX(
			stakerNode.UserPass,
			[]string{},
			"", // change addr
			stakerNode.XAddress,
			seedAmount-stakeAmount,
		)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX to xChainAddress %s", stakerNode.XAddress))
		}

		err = chainhelper.PChain().AwaitTransactionAcceptance(stakerNode.GetClient(), exportTxID, 30*time.Second)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to accept ExportTx: %s", exportTxID))
		}

		importTxID, err := stakerNode.GetClient().XChainAPI().ImportAVAX(
			stakerNode.UserPass,
			stakerNode.XAddress,
			constants.PlatformChainID.String())
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to import AVAX to xChainAddress %s", stakerNode.XAddress))
		}

		err = chainhelper.XChain().AwaitTransactionAcceptance(stakerNode.GetClient(), importTxID, 30*time.Second)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to wait for acceptance of transaction on XChain."))
		}

		err = chainhelper.PChain().CheckBalance(stakerNode.GetClient(), stakerNode.PAddress, 0)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected P Chain Balance after P -> X Transfer."))
		}

		err = chainhelper.XChain().CheckBalance(stakerNode.GetClient(), stakerNode.XAddress, "AVAX", seedAmount-stakeAmount)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer."))
		}
		logrus.Infof("Transferred leftover staker funds back to X Chain and verified X and P balances.")

		delegatorNode := topology.Node("delegator-node")

		exportTxID, err = delegatorNode.GetClient().PChainAPI().ExportAVAX(
			delegatorNode.UserPass,
			[]string{},
			"", // change addr
			delegatorNode.XAddress,
			seedAmount-stakeAmount,
		)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX to xChainAddress %s", delegatorNode.XAddress))
		}

		err = chainhelper.PChain().AwaitTransactionAcceptance(delegatorNode.GetClient(), exportTxID, 30*time.Second)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to accept ExportTx: %s", exportTxID))
		}

		txID, err := delegatorNode.GetClient().XChainAPI().ImportAVAX(
			delegatorNode.UserPass,
			delegatorNode.XAddress,
			constants.PlatformChainID.String(),
		)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX to xChainAddress %s", delegatorNode.XAddress))
		}

		err = chainhelper.XChain().AwaitTransactionAcceptance(delegatorNode.GetClient(), txID, 30*time.Second)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to Accept ImportTx: %s", importTxID))
		}

		err = chainhelper.PChain().CheckBalance(delegatorNode.GetClient(), delegatorNode.PAddress, 0)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected P Chain Balance after P -> X Transfer."))
		}

		err = chainhelper.XChain().CheckBalance(delegatorNode.GetClient(), delegatorNode.XAddress, "AVAX", seedAmount-stakeAmount)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer."))
		}

		logrus.Infof("Transferred leftover delegator funds back to X Chain and verified X and P balances.")
	}

	return testrunner.NewTestRunner("Derp1", testNetwork.Generate, test, timeout, timeoutBuffer)
}
