package ava_testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	// If we want, we can parameterize these tests so that they can be NNodeGeckoNetwork
	result["tenNodeBasicTest"] = TenNodeGeckoNetworkBasicTest{}
	result["tenNodeGetValidatorsTest"] = TenNodeNetworkGetValidatorsTest{}
	result["singleNodeBasicTest"] = SingleNodeGeckoNetworkBasicTest{}
	result["singleNodeGetValidatorsTest"] = SingleNodeNetworkGetValidatorsTest{}
	result["fiveStakingNodeGetValidatorsTest"] = FiveNodeStakingNetworkGetValidatorsTest{}
	result["fiveStakingNodeFullyConnectedTest"] = FiveNodeStakingNetworkFullyConnectedTest{}

	return result
}

