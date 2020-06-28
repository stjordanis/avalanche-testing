package gecko_client

import (
	"encoding/json"
	"github.com/palantir/stacktrace"
)

const (
	infoEndpoint = "ext/info"
)

type InfoApi struct {
	rpcRequester jsonRpcRequester
}

func (api InfoApi) GetPeers() ([]Peer, error) {
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(infoEndpoint, "info.peers", make(map[string]interface{}))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	var response GetPeersResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Peers, nil
}

func (api InfoApi) GetNodeId() (string, error) {
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(infoEndpoint, "info.getNodeID", make(map[string]interface{}))
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	var response GetNodeIDResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.NodeID, nil
}

