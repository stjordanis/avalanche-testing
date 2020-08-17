package ava_services

import (
	"github.com/docker/go-connections/nat"
)

/*
An implementation of AvaService representing the interactions available with a Gecko client of the Avalanche network
*/
type GeckoService struct {
	ipAddr      string
	stakingPort nat.Port
	jsonRpcPort nat.Port
}

func (g GeckoService) GetStakingSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, g.stakingPort)
}

func (g GeckoService) GetJsonRpcSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, g.jsonRpcPort)
}
