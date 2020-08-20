package services

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis/info"
	"github.com/ava-labs/avalanche-e2e-tests/utils/constants"
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

// NewAvalancheServiceAvailabilityChecker returns a new services.ServiceAvailabilityCheckerCore to
// check if an AvalancheService is ready
func NewAvalancheServiceAvailabilityChecker(timeout time.Duration) services.ServiceAvailabilityCheckerCore {
	return &GeckoServiceAvailabilityCheckerCore{
		timeout: timeout,
	}
}

// GeckoServiceAvailabilityCheckerCore implements services.ServiceAvailabilityCheckerCore
// that defines the criteria for a Gecko service being available
type GeckoServiceAvailabilityCheckerCore struct {
	timeout time.Duration
}

// IsServiceUp implements services.ServiceAvailabilityCheckerCore#IsServiceUp
// and returns true when the Gecko healthcheck reports that the node is available
func (g GeckoServiceAvailabilityCheckerCore) IsServiceUp(toCheck services.Service, dependencies []services.Service) bool {
	// NOTE: we don't check the dependencies intentionally, because we don't need to - a Gecko service won't report itself
	//  as up until its bootstrappers are up

	castedService := toCheck.(GeckoService)
	jsonRPCSocket := castedService.GetJSONRPCSocket()
	uri := fmt.Sprintf("http://%s:%d", jsonRPCSocket.GetIpAddr(), jsonRPCSocket.GetPort().Int())
	client := info.NewClient(uri, constants.DefaultRequestTimeout)

	if bootstrapped, err := client.IsBootstrapped("P"); err != nil || !bootstrapped {
		return false
	}
	if bootstrapped, err := client.IsBootstrapped("C"); err != nil || !bootstrapped {
		return false
	}
	if bootstrapped, err := client.IsBootstrapped("X"); err != nil || !bootstrapped {
		return false
	}

	time.Sleep(5 * time.Second)
	return true
}

// GetTimeout implements services.AvailabilityCheckerCore
func (g GeckoServiceAvailabilityCheckerCore) GetTimeout() time.Duration {
	return 90 * time.Second
}
