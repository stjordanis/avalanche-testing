package ava_testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	// TODO these should be parameterized
	result["tenNodeBasicTest"] = TenNodeGeckoNetworkBasicTest{}
	result["tenNodeGetValidatorsTest"] = TenNodeNetworkGetValidatorsTest{}
	result["singleNodeBasicTest"] = SingleNodeGeckoNetworkBasicTest{}
	result["singleNodeGetValidatorsTest"] = SingleNodeNetworkGetValidatorsTest{}

	result["fiveStakingNodeGetValidatorsTest"] = FiveNodeStakingNetworkGetValidatorsTest{}

	return result
}

