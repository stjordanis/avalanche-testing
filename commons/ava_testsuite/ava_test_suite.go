package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite/byzantine_tests"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {
	ChitSpammerImageName string
	NormalImageName string
}

func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	if a.ChitSpammerImageName != "" {
		result["stakingNetworkChitSpammerTest"] = byzantine_tests.StakingNetworkUnrequestedChitSpammerTest{
			UnrequestedChitSpammerImageName: a.ChitSpammerImageName,
			NormalImageName: a.NormalImageName,}
	}
	result["stakingNetworkFullyConnectedTest"] = StakingNetworkFullyConnectedTest{a.NormalImageName}
	result["stakingNetworkDuplicateNodeIdTest"] = StakingNetworkDuplicateNodeIdTest{a.NormalImageName}
	result["stakingNetworkRpcWorkflowTest"] = StakingNetworkRpcWorkflowTest{a.NormalImageName}

	return result
}

