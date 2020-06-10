package gecko_client

import (
	"github.com/docker/go-connections/nat"
)


type GeckoClient struct {
	pChainApi PChainApi
	adminApi  AdminApi
	healthApi HealthApi
}

func NewGeckoClient(ipAddr string, port nat.Port) *GeckoClient {
	rpcRequester := geckoJsonRpcRequester{
		ipAddr: ipAddr,
		port:   port,
	}

	return &GeckoClient{
		pChainApi: PChainApi{rpcRequester: rpcRequester},
		adminApi: AdminApi{rpcRequester: rpcRequester},
		healthApi: HealthApi{rpcRequester: rpcRequester},
	}
}

func (client GeckoClient) PChainApi() PChainApi {
	return client.pChainApi
}

func (client GeckoClient) AdminApi() AdminApi {
	return client.adminApi
}

func (client GeckoClient) HealthApi() HealthApi {
	return client.healthApi
}
