package gecko_client

import (
	"bytes"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

// ============= RPC Requester ===================
const (
	JSON_RPC_VERSION = "2.0"
)

type geckoJsonRpcRequester struct {
	ipAddr string
	port nat.Port
}

func (requester geckoJsonRpcRequester) makeRpcRequest(endpoint string, method string) ([]byte, error) {
	// Either Golang or Ava have a very nasty & subtle behaviour where duplicated '//' in the URL is treated as GET, even if it's POST
	// https://stackoverflow.com/questions/23463601/why-golang-treats-my-post-request-as-a-get-one
	endpoint = strings.TrimLeft(endpoint, "/")

	requestBodyStr := fmt.Sprintf(
		`{"jsonrpc": "%v", "method": "%v", "params":{},"id": %v}`,
		JSON_RPC_VERSION,
		method,
		// TODO let the user set this?
		1)
	requestBodyBytes := []byte(requestBodyStr)
	url := fmt.Sprintf("http://%v:%v/%v", requester.ipAddr, requester.port.Int(), endpoint)

	logrus.Tracef("Making request to url: %v", url)
	logrus.Tracef("Request body: %v", requestBodyStr)
	resp, err := http.Post(
		url,
		"application/json",
		bytes.NewBuffer(requestBodyBytes),
	)
	defer resp.Body.Close()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred when making JSON RPC POST request to %v", url)
	}
	statusCode := resp.StatusCode
	logrus.Tracef("Got response with status code: %v", statusCode)

	responseBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred when reading response body")
	}
	logrus.Tracef("Response body: %v", string(responseBodyBytes))

	if statusCode != 200 {
		return nil, stacktrace.NewError(
			"Received response with non-200 code '%v' and response body '%v'",
			statusCode,
			string(responseBodyBytes))
	}
	return responseBodyBytes, nil
}

// ============= Gecko Client ===================
type GeckoClient struct {
	pChainApi PChainApi
	adminApi  AdminApi
}

func NewGeckoClient(ipAddr string, port nat.Port) *GeckoClient {
	rpcRequester := geckoJsonRpcRequester{
		ipAddr: ipAddr,
		port:   port,
	}

	return &GeckoClient{
		pChainApi: PChainApi{rpcRequester: rpcRequester},
		adminApi: AdminApi{rpcRequester: rpcRequester},
	}
}

func (client GeckoClient) PChainApi() PChainApi {
	return client.pChainApi
}

func (client GeckoClient) AdminApi() AdminApi {
	return client.adminApi
}
