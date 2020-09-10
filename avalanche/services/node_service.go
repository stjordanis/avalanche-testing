package services

import (
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

// NodeService implements the Kurtosis generic services.Service interface that represents the minimum interface for a
// validator node
type NodeService interface {
	services.Service

	// GetStakingSocket returns the socket used for communication between nodes on the network
	GetStakingSocket() ServiceSocket
}
