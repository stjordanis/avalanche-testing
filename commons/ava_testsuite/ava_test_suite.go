package ava_testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {
	ChitSpammerImageName string
}

func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	result["stakingNetworkChitSpammerTest"] = StakingNetworkUnrequestedChitSpammerTest{&a.ChitSpammerImageName}
	result["stakingNetworkFullyConnectedTest"] = StakingNetworkFullyConnectedTest{}
	result["stakingNetworkDuplicateNodeIdTest"] = StakingNetworkDuplicateNodeIdTest{}
	result["stakingNetworkRpcWorkflowTest"] = StakingNetworkRpcWorkflowTest{}

	return result
}

