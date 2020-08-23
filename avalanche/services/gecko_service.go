package services

import (
	"github.com/docker/go-connections/nat"
)

// GeckoService implements AvalancheService
type GeckoService struct {
	ipAddr      string
	stakingPort nat.Port
	jsonRPCPort nat.Port
}

// GetStakingSocket implements AvalancheService
func (g GeckoService) GetStakingSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, g.stakingPort)
}

// GetJSONRPCSocket implements AvalancheService
func (g GeckoService) GetJSONRPCSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, g.jsonRPCPort)
}
