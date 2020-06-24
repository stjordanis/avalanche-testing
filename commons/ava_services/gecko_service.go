package ava_services

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"strconv"
)

type GeckoService struct {
	ipAddr string
}

func (g GeckoService) GetStakingSocket() services.ServiceSocket {
	stakingPort, err := nat.NewPort("tcp", strconv.Itoa(stakingPort))
	if err != nil {
		// Realllllly don't think we should deal with propagating this one.... it means the user mistyped an integer
		panic(err)
	}
	return *services.NewServiceSocket(g.ipAddr, stakingPort)
}

func (g GeckoService) GetJsonRpcSocket() services.ServiceSocket {
	httpPort, err := nat.NewPort("tcp", strconv.Itoa(httpPort))
	if err != nil {
		panic(err)
	}
	return *services.NewServiceSocket(g.ipAddr, httpPort)
}
