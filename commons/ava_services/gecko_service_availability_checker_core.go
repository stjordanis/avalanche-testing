package ava_services

import (
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	pChainId = "P"
	xChainId = "X"
)

type GeckoServiceAvailabilityCheckerCore struct {}
func (g GeckoServiceAvailabilityCheckerCore) IsServiceUp(toCheck services.Service, dependencies []services.Service) bool {
	castedService := toCheck.(GeckoService)
	jsonRpcSocket := castedService.GetJsonRpcSocket()
	client := gecko_client.NewGeckoClient(jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort())

	// We use the design pattern of a for-loop with return statements to avoid making unnecessary requests to the Gecko node
	//  by aborting early
	for _, chainId := range []string{pChainId, xChainId} {
		isBootstrapped, err := client.InfoApi().IsBootstrapped(chainId)
		if err != nil {
			logrus.Tracef("%s-Chain is not available due to error: %s", chainId, err.Error())
			return false
		}
		if !isBootstrapped {
			logrus.Tracef("%s-Chain is not available due not being bootstrapped yet", chainId)
			return false
		}
	}

	logrus.Trace("All chains are now bootstrapped successfully")
	return true
}

func (g GeckoServiceAvailabilityCheckerCore) GetTimeout() time.Duration {
	return 30 * time.Second
}


