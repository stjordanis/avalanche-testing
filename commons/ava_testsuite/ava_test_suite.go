package ava_testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	result["fiveStakingNodeGetValidatorsTest"] = FiveNodeStakingNetworkGetValidatorsTest{}
	result["fiveStakingNodeFullyConnectedTest"] = FiveNodeStakingNetworkFullyConnectedTest{}
	result["fiveStakingNodeDuplicateNodeIdTest"] = FiveNodeStakingNetworkDuplicateIdTest{}
	result["stakingNodeRpcWorkflowTest"] = StakingNetworkRpcWorkflowTest{}

	return result
}

