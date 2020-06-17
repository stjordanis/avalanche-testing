package gecko_client

import (
	"encoding/json"
	"github.com/palantir/stacktrace"
)

const (
	healthApiEndpoint = "ext/health"
)


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


