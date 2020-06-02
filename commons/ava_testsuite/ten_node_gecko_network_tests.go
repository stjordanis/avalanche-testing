package ava_testsuite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_networks"
	"github.com/kurtosis-tech/kurtosis/controller"
	"io/ioutil"
	"net/http"
)

type TenNodeGeckoNetworkBasicTest struct {}
func (s TenNodeGeckoNetworkBasicTest) Run(network interface{}, context controller.TestContext) {
	castedNetwork := network.(ava_networks.TenNodeGeckoNetwork)

	service, err := castedNetwork.GetGeckoService(0)
	if err != nil {
		context.Fatal(err)
	}
	httpSocket := service.GetJsonRpcSocket()

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


