package ava_testsuite

import (
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks/fixed_gecko_network"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

// =============== Basic Test ==================================
type SingleNodeGeckoNetworkBasicTest struct {}
func (test SingleNodeGeckoNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(fixed_gecko_network.FixedGeckoNetwork)

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
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(1, 1, false)
}

func (test SingleNodeGeckoNetworkBasicTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

// =============== Get Validators Test ==================================
type SingleNodeNetworkGetValidatorsTest struct{}
func (test SingleNodeNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(fixed_gecko_network.FixedGeckoNetwork)

	client, err := castedNetwork.GetGeckoClient(0)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get client"))
	}

	// TODO This retry logic is only necessary because there's not a way for Ava nodes to reliably report
	//  bootstrapping as complete; remove it when Gecko can report successful bootstrapping
	var validators []gecko_client.Validator
	for i := 0; i < 5; i++ {
		validators, err = client.PChainApi().GetCurrentValidators(nil)
		if err == nil {
			break
		}
		logrus.Error(stacktrace.Propagate(err, "Could not get current validators; sleeping for 5 seconds..."))
		time.Sleep(5 * time.Second)
	}
	// TODO This should go away as soon as Ava can reliably report bootstrapping as complete
	if validators == nil {
		context.Fatal(stacktrace.NewError("Could not get validators even after retrying!"))
	}

	for _, validator := range validators {
		logrus.Infof("Validator ID: %s", validator.Id)
	}
	// TODO figure out exactly how many validators this should actually have and set the value appropriately!
	context.AssertTrue(len(validators) >= 1)
}

func (test SingleNodeNetworkGetValidatorsTest) GetNetworkLoader() (testsuite.TestNetworkLoader, error) {
	return fixed_gecko_network.NewFixedGeckoNetworkLoader(1, 1, false)
}

func (test SingleNodeNetworkGetValidatorsTest) GetTimeout() time.Duration {
	return 30 * time.Second
}

