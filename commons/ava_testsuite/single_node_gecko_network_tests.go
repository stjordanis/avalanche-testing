package ava_testsuite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

const (
)

type SingleNodeGeckoNetworkBasicTest struct {}
func (test SingleNodeGeckoNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.SingleNodeGeckoNetwork)
	httpSocket := castedNetwork.GetNode().GetJsonRpcSocket()

	requestBody, err := json.Marshal(map[string]string{
		"jsonrpc": "2.0",
		"id": "1",
		"method": "admin.peers",
	})
	if err != nil {
		context.Fatal(err)
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%v:%v/ext/admin", httpSocket.GetIpAddr(), httpSocket.GetPort().Int()),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		context.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		context.Fatal(err)
	}

	// TODO parse the response as JSON and assert that we get the expected number of peers
	println(string(body))
}

func (test SingleNodeGeckoNetworkBasicTest) GetNetworkLoader() testsuite.TestNetworkLoader {
	return ava_networks.SingleNodeGeckoNetworkLoader{}
}

type SingleNodeNetworkGetValidatorsTest struct{}
func (test SingleNodeNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.SingleNodeGeckoNetwork)

	// TODO Move these into a better location
	RPC_BODY := `{"jsonrpc": "2.0", "method": "platform.getCurrentValidators", "params":{},"id": 1}`

	// Run RPC Test on PChain.
	var jsonStr = []byte(RPC_BODY)
	var jsonBuffer = bytes.NewBuffer(jsonStr)
	logrus.Infof("Test request as string: %s", jsonBuffer.String())

	var validatorList ValidatorList
	jsonRpcSocket := castedNetwork.GetNode().GetJsonRpcSocket()
	endpoint := fmt.Sprintf("http://%v:%v/%v", jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort().Int(), GetPChainEndpoint())
	resp, err := http.Post(endpoint, "application/json", jsonBuffer)
	if err != nil {
		context.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Fatalln(err)
	}

	var validatorResponse ValidatorResponse
	json.Unmarshal(body, &validatorResponse)

	validatorList = validatorResponse.Result["validators"]
	for _, validator := range validatorList {
		logrus.Infof("Validator id: %s", validator.Id)
	}
	context.AssertTrue(len(validatorList) >= 1)
}

func (test SingleNodeNetworkGetValidatorsTest) GetNetworkLoader() testsuite.TestNetworkLoader {
	return ava_networks.SingleNodeGeckoNetworkLoader{}
}

