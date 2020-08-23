package services

import "github.com/docker/go-connections/nat"

// ServiceSocket ...
type ServiceSocket struct {
	ipAddr string
	port   nat.Port
}

// NewServiceSocket ...
func NewServiceSocket(ipAddr string, port nat.Port) *ServiceSocket {
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
func (socket *ServiceSocket) GetPort() nat.Port {
	return socket.port
}
