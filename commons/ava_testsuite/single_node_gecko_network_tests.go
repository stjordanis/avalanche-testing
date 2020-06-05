package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
)

// =============== Basic Test ==================================
type SingleNodeGeckoNetworkBasicTest struct {}
func (test SingleNodeGeckoNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.NNodeGeckoNetwork)

	client, err := castedNetwork.GetGeckoClient(0)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get client"))
	}

	peers, err := client.AdminApi().GetPeers()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get peers"))
	}

	context.AssertTrue(len(peers) == 0)
}
func (test SingleNodeGeckoNetworkBasicTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return ava_networks.NewNNodeGeckoNetworkLoader(1, 1)
}

// =============== Get Validators Test ==================================
type SingleNodeNetworkGetValidatorsTest struct{}
func (test SingleNodeNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.NNodeGeckoNetwork)

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
	// TODO figure out exactly how many validators this should actually have and set the value appropriately!
	context.AssertTrue(len(validators) >= 1)
}

func (test SingleNodeNetworkGetValidatorsTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return ava_networks.NewNNodeGeckoNetworkLoader(1, 1)
}

