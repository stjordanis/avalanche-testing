package ava_services

import (
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

// AvaService implements the Kurtosis generic services.Service interface that represents the minimum interface an Avalanche node
type AvaService interface {
	services.Service

	// Gets the staking socket of an Avalanche node (which all nodes must have if they're part of an Avalanche network)
	GetStakingSocket() ServiceSocket
}
