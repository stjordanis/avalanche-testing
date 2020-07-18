package ava_services

import (
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

type GeckoServiceAvailabilityCheckerCore struct {}
func (g GeckoServiceAvailabilityCheckerCore) IsServiceUp(toCheck services.Service, dependencies []services.Service) bool {
	castedService := toCheck.(GeckoService)
	jsonRpcSocket := castedService.GetJsonRpcSocket()
	client := gecko_client.NewGeckoClient(jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort())
	healthInfo, err := client.HealthApi().GetLiveness()
	if err != nil {
		logrus.Trace(stacktrace.Propagate(err, "Error occurred in getting liveness info"))
		return false
	}

	// HACK HACK HACK we need to wait for bootstrapping to finish, and there is not API for this yet (in development)
	// TODO TODO TODO once bootstrapping checker is available, use that instead of just waiting
	if healthInfo.Healthy {
		time.Sleep(15 * time.Second)
	}
	return healthInfo.Healthy
}

func (g GeckoServiceAvailabilityCheckerCore) GetTimeout() time.Duration {
	return 90 * time.Second
}


