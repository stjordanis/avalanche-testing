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
	txFee             = 1 * units.Avax
)

const (
	validatorNodeName string = "validator-node"
	delegatorNodeName string = "delegator-node"
)

// Workflow test the workflow of booting/setting up validator and delegator nodes
// additionally it's useful for show casing the manual build the network and the topology
func Workflow(avalancheImage string) *testrunner.TestRunner {

	// create the nodes
	stakerNode := network.NewNode(validatorNodeName).
		Image(avalancheImage).
		SnowConf(2, 2)

	delegatorNode := network.NewNode(delegatorNodeName).
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
			AddNode(validatorNodeName, stakerUsername, stakerPassword).
			AddNode(delegatorNodeName, delegatorUsername, delegatorPassword).
			AddGenesis(validatorNodeName, genesisUsername, genesisPassword)

		// creates a genesis and funds the X addresses of the nodes
		topology.Genesis().
			FundXChainAddresses([]string{
				topology.Node(validatorNodeName).XAddress,
				topology.Node(delegatorNodeName).XAddress,
			},
				totalAmount,
			)

		// sets the nodes to validators and delegators
		// validatorNodeName - will have available after this op :
		// XChain - 10k - 5k - 2*txFee - 4998000000000
		// PChain - 5k - 3k = 2k
		topology.Node(validatorNodeName).BecomeValidator(totalAmount, seedAmount, stakeAmount, txFee)

		// delegatorNodeName - will have available after this op :
		// XChain - 10k - 5k - 2*txFee = 4998000000000
		// PChain - 5k - 3k = 2k
		topology.Node(delegatorNodeName).BecomeDelegator(totalAmount, seedAmount, stakeAmount, txFee, topology.Node(validatorNodeName).NodeID)

		// after setup we want to test moving amounts from P to X Chain and back
		stakerNode := topology.Node(validatorNodeName)

		// Lets move whats in the PChain back to the XChain = 2k - txFee (to be burned)
		exportTxID, err := stakerNode.GetClient().PChainAPI().ExportAVAX(
			stakerNode.UserPass,
			[]string{},
			"", // change addr
			stakerNode.XAddress,
			seedAmount-stakeAmount-txFee,
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

		// Now we should have
		// XChain: 10k - 5k - 2*txFee (1st op) = 4998000000000 + 3k - 2*txFee (2nd export)
		err = chainhelper.XChain().CheckBalance(stakerNode.GetClient(), stakerNode.XAddress, "AVAX",
			totalAmount-seedAmount-2*txFee+seedAmount-stakeAmount-2*txFee)
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
			seedAmount-stakeAmount-txFee,
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

		err = chainhelper.XChain().CheckBalance(delegatorNode.GetClient(), delegatorNode.XAddress, "AVAX",
			totalAmount-seedAmount-2*txFee+seedAmount-stakeAmount-2*txFee)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer."))
		}

		logrus.Infof("Transferred leftover delegator funds back to X Chain and verified X and P balances.")
	}

	return testrunner.NewTestRunner("Derp1", testNetwork.Generate, test, timeout, timeoutBuffer)
}
