package services

import (
	"github.com/kurtosis-tech/kurtosis-go/lib/services"
)

// ServiceSocket ...
type ServiceSocket struct {
	ipAddr string
	port   int
}

// NewServiceSocket ...
func NewServiceSocket(ipAddr string, port int) *ServiceSocket {
	return &ServiceSocket{
		ipAddr: ipAddr,
		port:   port,
	}
}

// GetIpAddr ...
func (socket *ServiceSocket) GetIpAddr() string {
	return socket.ipAddr
}

// GetPort ...
func (socket *ServiceSocket) GetPort() int {
	return socket.port
}

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

// NOTES: improved interface
// NodeService implements the Kurtosis generic services.Service interface that represents the minimum interface for a
// validator node
type NodeService interface {
	services.Service

	// GetStakingSocket returns the socket used for communication between nodes on the network
	GetStakingSocket() ServiceSocket
}
