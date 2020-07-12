package ava_testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {
	ChitSpammerImageName string
	NormalImageName string
}

func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	result["stakingNetworkChitSpammerTest"] = StakingNetworkUnrequestedChitSpammerTest{
		a.ChitSpammerImageName,
		              				a.NormalImageName,}
	result["stakingNetworkFullyConnectedTest"] = StakingNetworkFullyConnectedTest{a.NormalImageName}
	result["stakingNetworkDuplicateNodeIdTest"] = StakingNetworkDuplicateNodeIdTest{a.NormalImageName}
	result["stakingNetworkRpcWorkflowTest"] = StakingNetworkRpcWorkflowTest{a.NormalImageName}

	return result
}

