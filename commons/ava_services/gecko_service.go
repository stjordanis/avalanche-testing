package ava_services

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

type GeckoService struct {
	ipAddr string
	stakingPort nat.Port
	jsonRpcPort nat.Port
}

func (g GeckoService) GetStakingSocket() services.ServiceSocket {
	return *services.NewServiceSocket(g.ipAddr, g.stakingPort)
}

func (g GeckoService) GetJsonRpcSocket() services.ServiceSocket {
	return *services.NewServiceSocket(g.ipAddr, g.jsonRpcPort)
}
