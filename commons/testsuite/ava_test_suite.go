package testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) GetTests() map[string]testsuite.TestConfig {
	result := make(map[string]testsuite.TestConfig)

	singleNodeNetworkLoaer := networks.SingleNodeGeckoNetworkLoader{}

	/*
	result["singleNodeBasicTest"] = testsuite.TestConfig{
		Test: SingleNodeGeckoNetworkBasicTest{},
		NetworkLoader: singleNodeNetworkLoaer,
	}
	*/

	// TODO make make the network loader-getting step a part of the Test itself
	result["singleNodeGetValidatorsTest"] = testsuite.TestConfig{
		Test: SingleNodeNetworkGetValidatorsTest{},
		NetworkLoader: singleNodeNetworkLoaer,
	}

	return result
}

