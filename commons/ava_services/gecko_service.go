package ava_services

import (
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

type GeckoService struct {
	ipAddr string
}

func (g GeckoService) GetStakingSocket() services.ServiceSocket {
	return *services.NewServiceSocket(g.ipAddr, stakingPort)
}

func (g GeckoService) GetJsonRpcSocket() services.ServiceSocket {
	return *services.NewServiceSocket(g.ipAddr, httpPort)
}
