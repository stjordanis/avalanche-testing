package gecko_client

import (
	"encoding/json"
	"github.com/palantir/stacktrace"
)

const (
	adminEndpoint = "ext/admin"
)

type AdminApi struct {
	rpcRequester jsonRpcRequester
}

type NodeID struct {
	NodeID string `json:"nodeID"`
}

type Peer struct {
	IP string	`json:"ip"`
	PublicIP string 	`json:"publicIP"`
	Id string	`json:"id"`
	Version string	`json:"version"`
	LastSent string 	`json:"lastSent"`
	LastReceived string	`json:"lastReceived"`
}

type PeerList struct {
	Peers []Peer	`json:"peers"`
}

type GetPeersResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result PeerList	`json:"result"`
	Id int	`json:"id"`
}

type GetNodeIDResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result NodeID	`json:"result"`
	Id int	`json:"id"`
}

func (api AdminApi) GetPeers() ([]Peer, error) {
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(adminEndpoint, "admin.peers", make(map[string]interface{}))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response GetPeersResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Peers, nil
}

func (api AdminApi) GetNodeId() (string, error) {
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(adminEndpoint, "admin.getNodeID", make(map[string]interface{}))
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response GetNodeIDResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.NodeID, nil
}
