package services

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/info"
	"github.com/ava-labs/avalanche-testing/utils/constants"
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

// NewAvalancheServiceAvailabilityChecker returns a new services.ServiceAvailabilityCheckerCore to
// check if an AvalancheService is ready
func NewAvalancheServiceAvailabilityChecker(timeout time.Duration) services.ServiceAvailabilityCheckerCore {
	return &AvalancheServiceAvailabilityCheckerCore{
		timeout: timeout,
	}
}

// AvalancheServiceAvailabilityCheckerCore implements services.ServiceAvailabilityCheckerCore
// that defines the criteria for an Avalanche service being available
type AvalancheServiceAvailabilityCheckerCore struct {
	timeout                                                    time.Duration
	bootstrappedPChain, bootstrappedCChain, bootstrappedXChain bool
}

// IsServiceUp implements services.ServiceAvailabilityCheckerCore#IsServiceUp
// and returns true when the Avalanche healthcheck reports that the node is available
func (g AvalancheServiceAvailabilityCheckerCore) IsServiceUp(toCheck services.Service, dependencies []services.Service) bool {
	// NOTE: we don't check the dependencies intentionally, because we don't need to - an Avalanche service won't report itself
	//  as up until its bootstrappers are up

	castedService := toCheck.(AvalancheService)
	jsonRPCSocket := castedService.GetJSONRPCSocket()
	uri := fmt.Sprintf("http://%s:%d", jsonRPCSocket.GetIpAddr(), jsonRPCSocket.GetPort().Int())
	client := info.NewClient(uri, constants.DefaultRequestTimeout)

	if !g.bootstrappedPChain {
		if bootstrapped, err := client.IsBootstrapped("P"); err != nil || !bootstrapped {
			return false
		}
		g.bootstrappedPChain = true
	}

	if !g.bootstrappedCChain {
		if bootstrapped, err := client.IsBootstrapped("C"); err != nil || !bootstrapped {
			return false
		}
		g.bootstrappedCChain = true
	}

	if !g.bootstrappedXChain {
		if bootstrapped, err := client.IsBootstrapped("X"); err != nil || !bootstrapped {
			return false
		}
		g.bootstrappedXChain = true
	}

	time.Sleep(5 * time.Second)
	return true
}

// GetTimeout implements services.AvailabilityCheckerCore
func (g AvalancheServiceAvailabilityCheckerCore) GetTimeout() time.Duration {
	return 90 * time.Second
}
