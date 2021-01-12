package kurtosis

import (
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	"github.com/ava-labs/avalanche-testing/testsuite/tests/bombard"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/cchain"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/conflictvtx"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/connected"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/duplicate"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/managedasset"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/spamchits"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/workflow"
)

const (
	// The number of bits to make each test network, which dictates the max number of services a test can spin up
	// Here we choose 8 bits = 256 max services per test
	networkWidthBits = 8
)

// AvalancheTestSuite implements the Kurtosis TestSuite interface
type AvalancheTestSuite struct {
	ByzantineImageName string
	NormalImageName    string
}

// GetTests implements the Kurtosis TestSuite interface
func (a AvalancheTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	if a.ByzantineImageName != "" {
		result["chitSpammerTest"] = spamchits.NewChitSpammerTest(a.NormalImageName, a.ByzantineImageName)
		result["conflictingTxsVertexTest"] = conflictvtx.NewConflictingTxsVertexTest(a.NormalImageName, a.ByzantineImageName)
	}
	result["bombardXChainTest"] = bombard.NewBombardXChainTest(
		a.NormalImageName,
		1000,
		1000000,
		10*time.Second,
	)
	result["fullyConnectedNetworkTest"] = connected.NewFullyConnectedTest(a.NormalImageName, 70*time.Second)
	result["duplicateNodeIDTest"] = duplicate.NewDuplicateNodeIDTest(a.NormalImageName, nil)
	result["rpcWorkflowTest"] = workflow.NewRPCWorkflowTest(a.NormalImageName, nil)
	result["virtuousCorethTest"] = cchain.NewVirtuousCChainTest(a.NormalImageName, 100, 3, 1000000, 3*time.Second)
	result["managedAssetTest"] = managedasset.NewManagedAssetTest(a.NormalImageName)

	return result
}

// GetNetworkWidthBits returns the number of bits used to store the number of services
// the above comment on [networkWidthBits] explains this further
func (a AvalancheTestSuite) GetNetworkWidthBits() uint32 {
	return networkWidthBits
}
