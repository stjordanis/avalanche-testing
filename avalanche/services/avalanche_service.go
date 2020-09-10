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
func (g AvalancheService) GetStakingSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, g.stakingPort)
}

// GetJSONRPCSocket implements AvalancheService
func (g AvalancheService) GetJSONRPCSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, g.jsonRPCPort)
}
