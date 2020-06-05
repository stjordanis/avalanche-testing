package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// =============== Basic Test ==================================
type TenNodeGeckoNetworkBasicTest struct {}
func (s TenNodeGeckoNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.NNodeGeckoNetwork)

	// TODO check ALL nodes!
	client, err := castedNetwork.GetGeckoClient(0)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get client"))
	}

	peers, err := client.AdminApi().GetPeers()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get peers"))
	}

	context.AssertTrue(len(peers) == 9)
}

func (s TenNodeGeckoNetworkBasicTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return ava_networks.NewNNodeGeckoNetworkLoader(10, 3)
}


// =============== Get Validators Test ==================================
type TenNodeNetworkGetValidatorsTest struct{}
func (test TenNodeNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.NNodeGeckoNetwork)

	// TODO want to make sure all nodes agree about the number of validators!
	client, err := castedNetwork.GetGeckoClient(0)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get client"))
	}

	validators, err := client.PChainApi().GetCurrentValidators()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get current validators"))
	}

	for _, validator := range validators {
		logrus.Infof("Validator ID: %s", validator.Id)
	}
	// TODO change this to be specific
	context.AssertTrue(len(validators) >= 1)
}

func (test TenNodeNetworkGetValidatorsTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return ava_networks.NewNNodeGeckoNetworkLoader(10, 3)
}

