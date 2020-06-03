package ava_testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	/*
	result["singleNodeBasicTest"] = ava_testsuite.TestConfig{
		Test: SingleNodeGeckoNetworkBasicTest{},
		NetworkLoader: singleNodeNetworkLoaer,
	}
	*/

	// TODO make make the network loader-getting step a part of the Test itself
	result["singleNodeGetValidatorsTest"] = SingleNodeNetworkGetValidatorsTest{}

	return result
}

