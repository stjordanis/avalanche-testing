package services

import (
	"github.com/docker/go-connections/nat"
)

// AvalancheService implements AvalancheService
type AvalancheService struct {
	ipAddr      string
	stakingPort nat.Port
	jsonRPCPort nat.Port
}

// GetStakingSocket implements AvalancheService
func (service AvalancheService) GetStakingSocket() ServiceSocket {
	return *NewServiceSocket(service.ipAddr, service.stakingPort)
}

// GetJSONRPCSocket implements AvalancheService
func (service AvalancheService) GetJSONRPCSocket() ServiceSocket {
	return *NewServiceSocket(service.ipAddr, service.jsonRPCPort)
}
