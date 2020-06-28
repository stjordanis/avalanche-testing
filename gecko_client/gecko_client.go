package gecko_client

import (
	"github.com/docker/go-connections/nat"
)


type GeckoClient struct {
	pChainApi PChainApi
	xChainApi XChainApi
	adminApi  AdminApi
	healthApi HealthApi
	keystoreApi KeystoreApi
}

func NewGeckoClient(ipAddr string, port nat.Port) *GeckoClient {
	rpcRequester := geckoJsonRpcRequester{
		ipAddr: ipAddr,
		port:   port,
	}

	return clientFromRequester(rpcRequester)
}

// This method is exposed for mocking the Gecko client
func clientFromRequester(requester jsonRpcRequester) *GeckoClient {
	return &GeckoClient{
		pChainApi: PChainApi{rpcRequester: requester},
		xChainApi: XChainApi{rpcRequester: requester},
		adminApi: AdminApi{rpcRequester: requester},
		healthApi: HealthApi{rpcRequester: requester},
		keystoreApi: KeystoreApi{rpcRequester: requester},
	}
}

func (client GeckoClient) PChainApi() PChainApi {
	return client.pChainApi
}

func (client GeckoClient) XChainApi() XChainApi {
	return client.xChainApi
}

func (client GeckoClient) AdminApi() AdminApi {
	return client.adminApi
}

func (client GeckoClient) HealthApi() HealthApi {
	return client.healthApi
}

func (client GeckoClient) KeystoreApi() KeystoreApi {
	return client.keystoreApi
}
