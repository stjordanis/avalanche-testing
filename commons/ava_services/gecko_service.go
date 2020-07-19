package ava_services

type GeckoService struct {
	ipAddr string
}

func (g GeckoService) GetStakingSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, stakingPort)
}

func (g GeckoService) GetJsonRpcSocket() ServiceSocket {
	return *NewServiceSocket(g.ipAddr, httpPort)
}
