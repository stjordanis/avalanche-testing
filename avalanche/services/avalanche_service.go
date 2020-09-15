package services

// AvalancheService implements AvalancheService
type AvalancheService struct {
	ipAddr      string
	stakingPort int
	jsonRPCPort int
}

// GetStakingSocket implements AvalancheService
func (g AvalancheService) GetStakingSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, g.stakingPort)
}

// GetJSONRPCSocket implements AvalancheService
func (g AvalancheService) GetJSONRPCSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, g.jsonRPCPort)
}
