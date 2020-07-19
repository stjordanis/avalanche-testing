package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite/duplicate_node_id_test"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite/fully_connected_test"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite/rpc_workflow_test"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite/unrequested_chit_spammer_test"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite/verifier"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {
	ChitSpammerImageName string
	NormalImageName string
}

func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	if a.ChitSpammerImageName != "" {
		result["stakingNetworkChitSpammerTest"] = unrequested_chit_spammer_test.StakingNetworkUnrequestedChitSpammerTest{
			UnrequestedChitSpammerImageName: a.ChitSpammerImageName,
			NormalImageName: a.NormalImageName,
		}
	}
	result["stakingNetworkFullyConnectedTest"] = fully_connected_test.StakingNetworkFullyConnectedTest{
		ImageName: a.NormalImageName,
		Verifier: verifier.NetworkStateVerifier{},
	}
	result["stakingNetworkDuplicateNodeIdTest"] = duplicate_node_id_test.DuplicateNodeIdTest{
		ImageName: a.NormalImageName,
		Verifier: verifier.NetworkStateVerifier{},
	}
	result["stakingNetworkRpcWorkflowTest"] = rpc_workflow_test.StakingNetworkRpcWorkflowTest{
		ImageName: a.NormalImageName,
	}

	return result
}

