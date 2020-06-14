package gecko_client

import (
	"encoding/json"
	"github.com/palantir/stacktrace"
)

const (
	healthApiEndpoint = "ext/health"
)

type Check struct {
	Message map[string]interface{} 	`json:"message"`
	Timestamp string	`json:"timestamp"`
	Duration int 	`json:"duration"`
	ContiguousFailures int 	`json:"contiguousFailures"`
	TimeOfFirstFailure string	`json:"timeOfFirstFailure"`
}

type LivenessInfo struct {
	Checks map[string]Check	`json:"checks"`
	Healthy bool	`json:"healthy"`
}

type GetLivenessResponse struct {
	JsonRpcVersion string       `json:"jsonrpc"`
	Result         LivenessInfo `json:"result"`
	Id             int          `json:"id"`
}

type HealthApi struct {
	rpcRequester jsonRpcRequester
}

func (api HealthApi) GetLiveness() (LivenessInfo, error) {
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(healthApiEndpoint, "health.getLiveness", make(map[string]interface{}))
	if err != nil {
		return LivenessInfo{}, stacktrace.Propagate(err, "Error getting liveness")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response GetLivenessResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return LivenessInfo{}, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result, nil
}


