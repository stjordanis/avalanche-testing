package services

// AvalancheService implements AvalancheService
type AvalancheService struct {
	ipAddr      string
	stakingPort int
	jsonRPCPort int
}

// GetStakingSocket implements AvalancheService
func (service AvalancheService) GetStakingSocket() ServiceSocket {
	return *NewServiceSocket(service.ipAddr, service.stakingPort)
}

// GetJSONRPCSocket implements AvalancheService
func (service AvalancheService) GetJSONRPCSocket() ServiceSocket {
	return *NewServiceSocket(service.ipAddr, service.jsonRPCPort)
}
