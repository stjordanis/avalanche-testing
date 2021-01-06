package kurtosis

import (
	"time"

	"github.com/ava-labs/avalanche-testing/testsuite/tests/bombard"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/cchain"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/conflictvtx"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/connected"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/duplicate"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/spamchits"
	"github.com/ava-labs/avalanche-testing/testsuite/verifier"
	"github.com/ava-labs/avalanche-testing/testsuite_v2/tests"

	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"
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
		result["chitSpammerTest"] = spamchits.StakingNetworkUnrequestedChitSpammerTest{
			ByzantineImageName: a.ByzantineImageName,
			NormalImageName:    a.NormalImageName,
		}
		result["conflictingTxsVertexTest"] = conflictvtx.StakingNetworkConflictingTxsVertexTest{
			ByzantineImageName: a.ByzantineImageName,
			NormalImageName:    a.NormalImageName,
		}
	}
	result["bombardXChainTest"] = bombard.StakingNetworkBombardTest{
		ImageName:         a.NormalImageName,
		NumTxs:            1000,
		TxFee:             1000000,
		AcceptanceTimeout: 10 * time.Second,
	}
	result["fullyConnectedNetworkTest"] = connected.StakingNetworkFullyConnectedTest{
		ImageName: a.NormalImageName,
		Verifier:  verifier.NetworkStateVerifier{},
	}
	result["duplicateNodeIDTest"] = duplicate.DuplicateNodeIDTest{
		ImageName: a.NormalImageName,
		Verifier:  verifier.NetworkStateVerifier{},
	}

	result["virtuousCorethTest"] = cchain.NewVirtuousCChainTest(a.NormalImageName, 100, 3, 1000000, 3*time.Second)

	result["GetUTXOs"] = tests.GetUTXOs(a.NormalImageName)
	result["Workflow"] = tests.Workflow(a.NormalImageName)

	return result
}

// GetNetworkWidthBits returns the number of bits used to store the number of services
// the above comment on [networkWidthBits] explains this further
func (a AvalancheTestSuite) GetNetworkWidthBits() uint32 {
	return networkWidthBits
}
