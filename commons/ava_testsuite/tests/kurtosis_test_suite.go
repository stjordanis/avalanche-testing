package tests

import (
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/duplicate_node_id_test"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/fully_connected_test"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/rpc_workflow_test"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/unrequested_chit_spammer_test"
	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/verifier"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

/*
An implementation of testsuite.TestSuite for Ava
*/
type AvaTestSuite struct {
	ByzantineImageName string
	NormalImageName    string
}

/*
An implementation of testsuite.TestSuite that returns the tests registered with this test suite
*/
func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	if a.ByzantineImageName != "" {
		result["stakingNetworkChitSpammerTest"] = unrequested_chit_spammer_test.StakingNetworkUnrequestedChitSpammerTest{
			ByzantineImageName: a.ByzantineImageName,
			NormalImageName:    a.NormalImageName,
		}
		// This is better tested at unit test level
		// result["conflictingTxsVertexTest"] = conflicting_txs_vertex_test.StakingNetworkConflictingTxsVertexTest{
		// 	ByzantineImageName: a.ByzantineImageName,
		// 	NormalImageName:    a.NormalImageName,
		// }
	}
	result["stakingNetworkFullyConnectedTest"] = fully_connected_test.StakingNetworkFullyConnectedTest{
		ImageName: a.NormalImageName,
		Verifier:  verifier.NetworkStateVerifier{},
	}
	result["stakingNetworkDuplicateNodeIdTest"] = duplicate_node_id_test.DuplicateNodeIdTest{
		ImageName: a.NormalImageName,
		Verifier:  verifier.NetworkStateVerifier{},
	}
	result["stakingNetworkRpcWorkflowTest"] = rpc_workflow_test.StakingNetworkRpcWorkflowTest{
		ImageName: a.NormalImageName,
	}

	return result
}
