package tests

import (
	"time"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/avalanche-testing/testsuite_v2/builder/chainhelper"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"

	"github.com/ava-labs/avalanche-testing/testsuite_v2/scenarios"

	"github.com/ava-labs/avalanche-testing/testsuite_v2/builder/testrunner"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"
)

// GetUTXOs tests sending utxos send from each nodes to one node
func GetUTXOs(avalancheImage string) *testrunner.TestRunner {

	// sets up the network - uses a default setup network
	testNetwork := scenarios.NewFiveNetworkStaking(avalancheImage).NewNetwork()

	// test timeout - needs a few refactors to have a smaller timeout
	// like less nodes or go routine calls with async-safe clients
	timeout := 5 * time.Minute

	// No idea what this is
	timeoutBuffer := 5 * time.Minute

	// the actual test
	test := func(network networks.Network, context testsuite.TestContext) {

		// builds the topology of the test - uses a default
		topology := scenarios.NewFiveNetworkStaking(avalancheImage).NewTopology(network, &context)

		// all nodes were genesis'd with 10k AVAX
		// 5k in PChain ( 3k staking + 2k not locked)
		// 5K in the XChain
		sendingNodes := []string{"first", "second", "third", "fourth"}
		password := "MyNameIs!Jeff"
		receivingNode := topology.Node("fifth")

		var txsIDs []ids.ID
		for _, nodeName := range sendingNodes {
			for i := 0; i < 10; i++ {
				// send AVAX
				txID, err := topology.Node(nodeName).GetClient().XChainAPI().Send(api.UserPass{
					Username: nodeName,
					Password: password,
				},
					[]string{},
					"",
					1000*units.Avax,
					"AVAX",
					receivingNode.XAddress,
					"",
				)

				if err != nil {
					logrus.Info(stacktrace.Propagate(err, "Failed to send fund from %s to %s on the XChain.", nodeName, receivingNode.NodeID))
				}
				txsIDs = append(txsIDs, txID)
			}
		}

		for _, txID := range txsIDs {
			err := chainhelper.XChain().AwaitTransactionAcceptance(receivingNode.GetClient(), txID, 5*time.Second)
			if err != nil {
				context.Fatal(stacktrace.Propagate(err, "Unable to check transaction status"))
			}
		}

		err := chainhelper.XChain().CheckBalance(receivingNode.GetClient(), receivingNode.XAddress, "AVAX", (5+4)*units.KiloAvax)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected balance."))
		}

		lastAddr := ""
		lastIdx := ""
		utxoCount := 0

		for i := 0; i < 10; i++ {
			utxosBytes, index, err := receivingNode.GetClient().XChainAPI().GetUTXOs([]string{receivingNode.XAddress}, 4, lastAddr, lastIdx)
			if err != nil {
				context.Fatal(stacktrace.Propagate(err, "Unable to fetch UTXOs"))
			}

			lastAddr = index.Address
			lastIdx = index.UTXO

			utxoCount = +len(utxosBytes)
		}

		if utxoCount != 40 {
			context.Fatal(stacktrace.Propagate(err, "Unexpected number of UTXOs - expected: %d got %d", 40, utxoCount))
		}
	}

	return testrunner.NewTestRunner("GetUTXOs", testNetwork.Generate, test, timeout, timeoutBuffer)
}
