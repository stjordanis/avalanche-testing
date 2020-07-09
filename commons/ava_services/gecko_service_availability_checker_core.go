package ava_services

import (
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	AVAILABILITY_TRANSACTION    = "NO_VALUE_JUST_FOR_CHECKING_LIVENESS"
	AVAILABILITY_USER   = "NO_VALUE_JUST_FOR_CHECKING_LIVENESS"
	AVAILABILITY_PASSWORD   = "NO_VALUE_JUST_FOR_CHECKING_LIVENESS"
	API_NOT_AVAILABLE_ERROR_STR = "404 not found"
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

	// check if xChain API handler is up yet
	xchainAvailability := false
	_, err = client.XChainApi().GetTxStatus(AVAILABILITY_TRANSACTION)
	if err != nil {
		xchainAvailability = !strings.Contains(err.Error(), API_NOT_AVAILABLE_ERROR_STR)
	} else {
		xchainAvailability = true
	}

	// check if pChain API handler is up yet
	pchainAvailability := false
	_, err = client.PChainApi().GetCurrentValidators(nil)
	if err != nil {
		pchainAvailability = !strings.Contains(err.Error(), API_NOT_AVAILABLE_ERROR_STR)
	} else {
		pchainAvailability = true
	}

	// check if keyChain API handler is up yet
	keyChainAvailability := false
	_, err = client.KeystoreApi().CreateUser(AVAILABILITY_USER, AVAILABILITY_PASSWORD)
	if err != nil {
		keyChainAvailability = !strings.Contains(err.Error(), API_NOT_AVAILABLE_ERROR_STR)
	} else {
		keyChainAvailability = true
	}

	// check if info API handler is up yet
	infoAvailability := false
	_, err = client.InfoApi().GetNodeId()
	if err != nil {
		infoAvailability = !strings.Contains(err.Error(), API_NOT_AVAILABLE_ERROR_STR)
	} else {
		infoAvailability = true
	}

	return healthInfo.Healthy && xchainAvailability && pchainAvailability && keyChainAvailability && infoAvailability
}

func (g GeckoServiceAvailabilityCheckerCore) GetTimeout() time.Duration {
	return 90 * time.Second
}


