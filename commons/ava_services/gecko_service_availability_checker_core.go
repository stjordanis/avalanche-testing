package ava_services

import (
	"bytes"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type GeckoServiceAvailabilityCheckerCore struct {}
func (g GeckoServiceAvailabilityCheckerCore) IsServiceUp(toCheck services.Service, dependencies []services.Service) bool {
	castedService := toCheck.(GeckoService)
	httpSocket := castedService.GetJsonRpcSocket()

	jsonBodyBytes := []byte(CHECK_LIVENESS_RPC_BODY)
	jsonBodyBuffer := bytes.NewBuffer(jsonBodyBytes)
	resp, err := http.Post(
		fmt.Sprintf("http://%v:%v/ext/admin", httpSocket.GetIpAddr(), httpSocket.GetPort().Int()),
		"application/json",
		jsonBodyBuffer,
	)
	if err != nil {
		logrus.Tracef("Error occurred in request to get peers: %v", err)
		return false
	}
	defer resp.Body.Close()

	// TODO parse the response body as JSON and assert that we get the expected number of peers!!
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Tracef("Error occurred when parsing response body: %v", err)
		return false
	}

	return true
}

func (g GeckoServiceAvailabilityCheckerCore) GetTimeout() time.Duration {
	return 30 * time.Second
}


