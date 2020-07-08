package ava_testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

const (
	BYZANTINE_GECKO_IMAGE_NAME = "gecko-byzantine-634a4d0:latest"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)
	byzantineGeckoImageName := BYZANTINE_GECKO_IMAGE_NAME

	result["stakingNodeByzantineTest"] = StakingNetworkUnrequestedChitSpammerTest{&byzantineGeckoImageName}
	result["stakingNetworkFullyConnectedTest"] = StakingNetworkFullyConnectedTest{}
	result["stakingNetworkDuplicateNodeIdTest"] = StakingNetworkDuplicateNodeIdTest{}
	result["stakingNetworkRpcWorkflowTest"] = StakingNetworkRpcWorkflowTest{}

	return result
}

