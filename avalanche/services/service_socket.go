package services

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
