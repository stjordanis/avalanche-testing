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
	"time"
)


type FiveNodeStakingNetworkGetValidatorsTest struct{}
func (test FiveNodeStakingNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(ava_networks.FiveNodeStakingGeckoNetwork)

	// TODO Move these into a better location
	RPC_BODY := `{"jsonrpc": "2.0", "method": "platform.getCurrentValidators", "params":{},"id": 1}`
	RETRIES := 5

	// TODO we shouldn't need to retry once we wait for the network to come up
	RETRY_WAIT_SECONDS := 5*time.Second

	// Run RPC Test on PChain.
	var jsonStr = []byte(RPC_BODY)
	var jsonBuffer = bytes.NewBuffer(jsonStr)
	logrus.Infof("Test request as string: %s", jsonBuffer.String())

	var validatorList ValidatorList
	service, err := castedNetwork.GetGeckoService(0)
	if err != nil {
		panic(err)
	}
	jsonRpcSocket := service.GetJsonRpcSocket()
	endpoint := fmt.Sprintf("http://%v:%v/%v", jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort().Int(), GetPChainEndpoint())
	for i := 0; i < RETRIES; i++ {
		resp, err := http.Post(endpoint, "application/json", jsonBuffer)
		if err != nil {
			logrus.Infof("Attempted connection...: %s", err.Error())
			logrus.Infof("Could not connect on attempt %d, retrying...", i+1)
			time.Sleep(RETRY_WAIT_SECONDS)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Fatalln(err)
		}

		var validatorResponse ValidatorResponse
		json.Unmarshal(body, &validatorResponse)

		validatorList = validatorResponse.Result["validators"]
		if len(validatorList) > 0 {
			logrus.Infof("Found validators!")
			break
		}
	}
	for _, validator := range validatorList {
		logrus.Infof("Validator id: %s", validator.Id)
	}
	context.AssertTrue(len(validatorList) >= 1)
}

func (test FiveNodeStakingNetworkGetValidatorsTest) GetNetworkLoader() testsuite.TestNetworkLoader {
	return ava_networks.FiveNodeStakingGeckoNetworkLoader{}
}

