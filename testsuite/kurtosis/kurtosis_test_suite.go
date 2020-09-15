package kurtosis

import (
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"
	"time"

	"github.com/ava-labs/avalanche-testing/testsuite/tests/bombard"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/conflictvtx"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/connected"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/duplicate"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/spamchits"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/workflow"
	"github.com/ava-labs/avalanche-testing/testsuite/verifier"
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
		result["stakingNetworkChitSpammerTest"] = spamchits.StakingNetworkUnrequestedChitSpammerTest{
			ByzantineImageName: a.ByzantineImageName,
			NormalImageName:    a.NormalImageName,
		}
		result["conflictingTxsVertexTest"] = conflictvtx.StakingNetworkConflictingTxsVertexTest{
			ByzantineImageName: a.ByzantineImageName,
			NormalImageName:    a.NormalImageName,
		}
	}
	result["stakingNetworkBombardXChainTest"] = bombard.StakingNetworkBombardTest{
		ImageName:         a.NormalImageName,
		NumTxs:            1000,
		TxFee:             1000000,
		AcceptanceTimeout: 10 * time.Second,
	}
	result["stakingNetworkFullyConnectedTest"] = connected.StakingNetworkFullyConnectedTest{
		ImageName: a.NormalImageName,
		Verifier:  verifier.NetworkStateVerifier{},
	}
	result["stakingNetworkDuplicateNodeIDTest"] = duplicate.DuplicateNodeIDTest{
		ImageName: a.NormalImageName,
		Verifier:  verifier.NetworkStateVerifier{},
	}
	result["StakingNetworkRPCWorkflowTest"] = workflow.StakingNetworkRPCWorkflowTest{
		ImageName: a.NormalImageName,
	}

	return result
}

func (a AvalancheTestSuite) GetNetworkWidthBits() uint32 {
	return networkWidthBits
}

