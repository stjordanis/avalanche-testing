package gecko_client

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client/rpc_requester"
	"time"
)

const (
	requestTimeout = 10 * time.Second
)

type GeckoClient struct {
	pChainApi   PChainApi
	xChainApi   XChainApi
	infoApi     InfoApi
	healthApi   HealthApi
	keystoreApi KeystoreApi
}

func NewGeckoClient(ipAddr string, port nat.Port) *GeckoClient {
	rpcRequester := rpc_requester.NewGeckoJsonRpcRequester(ipAddr, port, requestTimeout)
	return clientFromRequester(rpcRequester)
}

// This method is exposed for mocking the Gecko client
func clientFromRequester(requester rpc_requester.JsonRpcRequester) *GeckoClient {
	return &GeckoClient{
		pChainApi:   PChainApi{rpcRequester: requester},
		xChainApi:   XChainApi{rpcRequester: requester},
		infoApi:     InfoApi{rpcRequester: requester},
		healthApi:   HealthApi{rpcRequester: requester},
		keystoreApi: KeystoreApi{rpcRequester: requester},
	}
}

func (client GeckoClient) PChainApi() PChainApi {
	return client.pChainApi
}

func (client GeckoClient) XChainApi() XChainApi {
	return client.xChainApi
}

func (client GeckoClient) InfoApi() InfoApi {
	return client.infoApi
}

func (client GeckoClient) HealthApi() HealthApi {
	return client.healthApi
}

func (client GeckoClient) KeystoreApi() KeystoreApi {
	return client.keystoreApi
}
